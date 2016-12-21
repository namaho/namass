package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var config Config

type SSCmdDataAddPort struct {
	Port     int    `json:"port"`
	Password string `json:"password"`
}

type SSCmdDataDelPort struct {
	Port int `json:"port"`
}

type Server struct {
	IP   string `json:"ip"`
	Area int    `json:"area"`
}

type ContextHandler interface {
	ServeHTTPContext(context.Context, http.ResponseWriter, *http.Request)
}

type ContextAdapter struct {
	ctx     context.Context
	handler ContextHandler
}

type ContextHandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

func (h ContextHandlerFunc) ServeHTTPContext(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	h(ctx, rw, req)
}

func (ca *ContextAdapter) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ca.handler.ServeHTTPContext(ca.ctx, rw, req)
}

type SockClient struct {
	conn *net.UnixConn
}

func (sc SockClient) Write(command string) {
	_, err := sc.conn.Write([]byte(command))
	if err != nil {
		log.Error(err)
	}
}

func (sc SockClient) Stop(command string) {
	sc.conn.Close()
	os.Remove(config.SSAgentUnixSock)
}

func NewSockClient() (SockClient, error) {
	sock_type := "unixgram"
	laddr := net.UnixAddr{config.SSAgentUnixSock, sock_type}
	conn, err := net.DialUnix(
		sock_type, &laddr,
		&net.UnixAddr{config.SSDaemonUnixSock, sock_type})
	if err != nil {
		panic(err)
	}

	go func(conn *net.UnixConn) {
		for {
			var buf [1024]byte
			n, err := conn.Read(buf[:])
			if err != nil {
				panic(err)
			}
			resp := string(buf[:n])
			if strings.HasPrefix(resp, "stat") {
				transferStat := resp[6:len(resp)]
				ReportTransfer(transferStat)
			} else {
				log.Debug(resp)
			}
		}
	}(conn)

	// TODO check error

	return SockClient{conn}, nil
}

func execCmd(ctx context.Context, cmd string) {
	log.Info("execute command " + cmd)
	sc := ctx.Value("SockClient").(SockClient)
	sc.Write(cmd)
}

func addPort(ctx context.Context, port, password string) {
	cmd := "add: {\"server_port\":" + port + ", \"password\":\"" + password + "\"}"
	execCmd(ctx, cmd)
}

func delPort(ctx context.Context, port string) {
	cmd := "remove: {\"server_port\":" + port + "}"
	execCmd(ctx, cmd)
}

func SSCmdAddPort(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var cmdData SSCmdDataAddPort
	err := decoder.Decode(&cmdData)
	if err != nil {
		log.Error(err)
	}

	addPort(ctx, strconv.Itoa(cmdData.Port), cmdData.Password)

	fmt.Fprintf(w, "ok")
}

func SSCmdDelPort(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var cmdData SSCmdDataDelPort
	err := decoder.Decode(&cmdData)
	if err != nil {
		log.Error(err)
	}

	delPort(ctx, strconv.Itoa(cmdData.Port))

	fmt.Fprintf(w, "ok")
}

func StartHTTPServerWithContext(ctx context.Context) {
	http.Handle("/ssagent/add_port", &ContextAdapter{ctx, ContextHandlerFunc(SSCmdAddPort)})
	http.Handle("/ssagent/del_port", &ContextAdapter{ctx, ContextHandlerFunc(SSCmdDelPort)})
	log.Info("Listening on " + config.Listen)
	http.ListenAndServe(config.Listen, nil)
}

func GetIPByInterfaceName(ifaceName string) (string, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return "", err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}

		if ip.To4() != nil {
			return ip.String(), nil
		}
	}

	return "", errors.New("no found")
}

func DoHttpRequest(url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return &http.Response{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("ServerToken", config.ServerToken)
	httpClient := &http.Client{}
	return httpClient.Do(req)
}

func ReportThisServer(ctx context.Context) {
	ip, err := GetIPByInterfaceName(config.ReportInterface)
	if err != nil {
		log.Error(err)
	}
	areaId, err := strconv.Atoi(config.Area)
	server := Server{IP: ip, Area: areaId}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(server)
	resp, err := DoHttpRequest("https://"+config.SSWebHTTPAddress+"/ssweb/server/discover", b)
	if err != nil {
		log.Error(err)
		return
	}

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
	}
	configMap := map[string]string{}
	err = json.Unmarshal(bs, &configMap)
	if err != nil {
		log.Panic(err)
	}
	for port, password := range configMap {
		go addPort(ctx, port, password)
	}
}

func ReportTransfer(transferStat string) {
	b := bytes.NewBuffer([]byte(transferStat))
	log.Debug("send transfer: " + transferStat)
	resp, err := DoHttpRequest("https://"+config.SSWebHTTPAddress+"/ssweb/transfer/report", b)
	if err != nil {
		log.Error(err)
		return
	}

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
	}

	log.Debug(string(bs))
}

func main() {
	var configFile string
	flag.StringVar(&configFile, "c", "ssagent.json", "configuration file")
	flag.Parse()

	configStr, err := ReadConfigFile(configFile)
	if err != nil {
		log.Panic(err)
	}

	err = json.Unmarshal(configStr, &config)
	if err != nil {
		log.Panic(err)
	}

	SetupLog()

	sc, err := NewSockClient()
	if err != nil {
		log.Error(err)
	}
	sc.Write("ping")

	ctx := context.WithValue(context.Background(), "SockClient", sc)

	ReportThisServer(ctx)

	StartHTTPServerWithContext(ctx)
}
