package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"encoding/json"
	"encoding/hex"
	"github.com/shopspring/decimal"
	"github.com/ugorji/go/codec"

	"github.com/nttdots/go-dots/dots_common/messages"
	"github.com/nttdots/go-dots/dots_client/task"
	"github.com/nttdots/go-dots/libcoap"
	log "github.com/sirupsen/logrus"
	common "github.com/nttdots/go-dots/dots_common"
	dots_config "github.com/nttdots/go-dots/dots_client/config"
	client_message "github.com/nttdots/go-dots/dots_client/messages"
	restful_router "github.com/nttdots/go-dots/dots_client/router"
)

const (
	DEFAULT_DOTS_SERVER_ADDRESS = "127.0.0.1"
)

var (
	server            string
	serverIP          net.IP
	signalChannelPort int
	dataChannelPort   int
	socket            string
	certFile          string
	clientCertFile    string
	clientKeyFile     string

	identity          string
	psk               string
	configFile        string
	defaultConfigFile = "dots_client.yaml"
)

func init() {
	abs, _ := filepath.Abs(os.Args[0])
	execDir := filepath.Dir(abs)
	certPath := getDefaultCertPath(execDir)
	defaultCertFile := filepath.Join(certPath, "ca-cert.pem")
	defaultClientCertFile := filepath.Join(certPath, "client-cert.pem")
	defaultClientKeyFile := filepath.Join(certPath, "client-key.pem")

	flag.StringVar(&server, "server", DEFAULT_DOTS_SERVER_ADDRESS, "dots Server address")
	flag.IntVar(&signalChannelPort, "signalChannelPort", common.DEFAULT_SIGNAL_CHANNEL_PORT, "dots signal channel Server port")
	flag.IntVar(&dataChannelPort, "dataChannelPort", common.DEFAULT_DATA_CHANNEL_PORT, "dots data channel Server port")
	flag.StringVar(&socket, "socket", common.DEFAULT_CLIENT_SOCKET_FILE, "dots client socket")
	flag.StringVar(&certFile, "certFile", defaultCertFile, "cert file path")
	flag.StringVar(&clientCertFile, "clientCertFile", defaultClientCertFile, "client cert file path")
	flag.StringVar(&clientKeyFile, "clientKeyFile", defaultClientKeyFile, "client key file path")

	flag.StringVar(&identity, "identity", "", "identity for DTLS PSK")
	flag.StringVar(&psk, "psk", "", "DTLS PSK")

	flag.StringVar(&configFile, "config", defaultConfigFile, "config yaml file")
}

// These variables hold the server connection configurations.
var signalChannelAddress string
var dataChannelAddress string

func connectSignalChannel(orgEnv *task.Env) (env *task.Env, err error) {
	var ctx *libcoap.Context
	var sess *libcoap.Session
	var oSess *libcoap.Session
	var addr libcoap.Address

	libcoap.Startup()

	addr, err = libcoap.AddressOf(serverIP, uint16(signalChannelPort))
	if err != nil {
		log.WithError(err).Error("AddressOf() failed")
		goto error
	}

	if 0 < len(psk) {
		log.WithField("identity", identity).WithField("psk", psk).Info("Using PSK")

		ctx = libcoap.NewContext(nil)
		if ctx == nil {
			log.Error("NewContext() -> nil")
			err = errors.New("NewContext() -> nil")
			goto error
		}

		sess = ctx.NewClientSessionPSK(addr, libcoap.ProtoDtls, identity, []byte(psk))
		if sess == nil {
			log.Error("NewClientSessionPSK() -> nil")
			err = errors.New("NewClientSessionPSK() -> nil")
			goto error
		}

	} else {
		dtlsParam := libcoap.DtlsParam { &certFile, nil, &clientCertFile, &clientKeyFile, config.PinnedCertificate }
		if orgEnv == nil {
			ctx = libcoap.NewContextDtls(nil, &dtlsParam, int(libcoap.CLIENT_PEER), nil)
			if ctx == nil {
				log.Error("NewContextDtls() -> nil")
				err = errors.New("NewContextDtls() -> nil")
				goto error
			}
		} else {
			ctx = orgEnv.CoapContext()
		}

		sess = ctx.NewClientSessionDTLS(addr, libcoap.ProtoDtls)
		if sess == nil {
			if orgEnv == nil {
				log.Error("NewClientSessionDTLS() -> nil")
				err = errors.New("NewClientSessionDTLS() -> nil")
				goto error
			} else {
				log.Debug("NewClientSessionDTLS() -> nil. Retry re-create new DTLS session")
				connectSignalChannel(orgEnv)
			}
		}
	}
	// create resource for heartbeat mechanism from server
	if ctx != nil {
		resource := libcoap.ResourceUnknownInit()
		ctx.AddResource(resource)
		resource.RegisterServerHandler(libcoap.RequestPut, heartbeatHandler())
	}

	if (orgEnv == nil){
		env = task.NewEnv(ctx, sess)
	} else {
		oSess = orgEnv.CoapSession()
		env = orgEnv
	}

	ctx.RegisterEventHandler(func( session *libcoap.Session, event libcoap.Event){
		if event == libcoap.EventSessionConnected {
			if orgEnv != nil {
				orgEnv.SetReplacingSession(session)
				env.SetIsServerStopped(false)
			}
		} else if event == libcoap.EventSessionDisconnected {
			if orgEnv == nil {
				log.Warn("Session is disconnected.")
				env.SetIsServerStopped(true)
				return
			}
			session.SessionRelease()
			restartConnection(env)
		} else if event == libcoap.EventPartialBlock {
			log.Debugf("Received Partial Block")
			env.SetIsPartialBlock(true)
		}
	})

	ctx.RegisterResponseHandler(func(_ *libcoap.Context, _ *libcoap.Session, _ *libcoap.Pdu, received *libcoap.Pdu) {
		handleResponse(env, received)
		if received != nil && oSess != nil && oSess == env.CoapSession(){
			sess.SessionRelease()
			log.Debugf("Restarted connection successfully with current session: %+v.", oSess.String())
			env.Run(task.NewHeartBeatTask(
					time.Duration(config.DefaultSessionConfiguration.HeartbeatInterval) * time.Second,
					config.DefaultSessionConfiguration.MissingHbAllowed,
					time.Duration(config.DefaultSessionConfiguration.AckTimeout)* time.Second,
					heartbeatResponseHandler,
					heartbeatTimeoutHandler))
		}
	})

	ctx.RegisterNackHandler(func(_ *libcoap.Session, sent *libcoap.Pdu, reason libcoap.NackReason) {
		if (reason == libcoap.NackRst){
			// Pong message
			handleResponse(env, sent)
		} else if (reason == libcoap.NackTooManyRetries){
			// Ping timeout
			handleRequestTimeout(env, sent)
		} else {
			// Unsupported type
			log.Infof("nack_handler gets fired with unsupported reason type : %+v.", reason)
		}
	})
	return

error:
	cleanupSignalChannel(ctx, sess)
	return
}

// heartbeat handler
func heartbeatHandler() libcoap.MethodHandler {
	return func(ctx *libcoap.Context, rsrc *libcoap.Resource, sess *libcoap.Session, request *libcoap.Pdu, token *[]byte, query *string, response *libcoap.Pdu) {
		log.Info("Handle receive heartbeat from server")
		log.Debugf("request.Data=\n%s", hex.Dump(request.Data))
		// Decode heartbeat message
		dec := codec.NewDecoder(bytes.NewReader(request.Data), common.NewCborHandle())
		var v messages.HeartBeatRequest
		err := dec.Decode(&v)
		if err != nil {
			log.WithError(err).Warn("CBOR Decode failed.")
			return
		}
		log.Infof("        CBOR decoded: %+v", v.String())
		body, errMsg := messages.ValidateHeartBeatMechanism(request)
		response.MessageID = request.MessageID
        response.Token     = request.Token
		if body == nil && errMsg != "" {
			log.Error(errMsg)
			response.Code = libcoap.ResponseInternalServerError
			response.Type = libcoap.TypeNon
			response.Data = []byte(errMsg)
		} else if body != nil && errMsg != "" {
			log.Error(errMsg)
			response.Code = libcoap.ResponseBadRequest
			response.Type = libcoap.TypeNon
			response.Data = []byte(errMsg)
		} else {
			response.Code = libcoap.ResponseChanged
			response.Type = libcoap.TypeNon
		}
		log.Debugf("response=%+v", response)
		task.SetIsReceiveHeartBeat(true)
		return
	}
}

func cleanupSignalChannel(ctx *libcoap.Context, sess *libcoap.Session) {
	if ctx != nil {
		ctx.FreeContext()
	}
	libcoap.Cleanup()
}

/*
 * serverHandler is a request handler function to the servers.
 */
func makeServerHandler(env *task.Env) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		// _, requestName := path.Split(r.URL.Path)
		// Split requestName and QueryParam
		var tmpPaths []string
		paths := strings.Split(r.URL.Path, "?")
		if len(paths) > 1 {
			tmpPaths = strings.Split(paths[0], "/")
			queryPaths := strings.Split(paths[1],"&")
			if len(queryPaths) > 1 {
				tmpPaths = append(tmpPaths, queryPaths...)
			} else {
				tmpPaths = append(tmpPaths, paths[1])
			}
		} else {
			tmpPaths = strings.Split(r.URL.Path, "/")
		}
		var requestName = ""
		var tmpPath string
		var requestQuerys []string
		for i := len(tmpPaths) - 1; i >=0; i-- {
			tmpPath = tmpPaths[i]
			// if include =, use for QueryParam and check previous path
			if strings.Contains(tmpPath, "=") {
				continue
			}
			requestName = tmpPath
			requestQuerys = tmpPaths[i+1:]
			break
		}

		// The 'cuid' should be less than or equal to 22 characters
		for _,query := range requestQuerys {
			querySplit := strings.Split(query, "cuid=")
			if len(querySplit) > 1 && len(querySplit[1]) > int(messages.CUID_LEN) {
				log.Warnf("The 'cuid' (%+v) should not be greater than 22 characters", len(querySplit[1]))
				return
			}
		}

		options := make(map[messages.Option]string)
		// Create observe option
		observeStr := r.Header.Get(string(messages.OBSERVE))
		if observeStr != "" {
			options[messages.OBSERVE] = observeStr
		}
		// Create If-Match option
		if val, ok := r.Header[string(messages.IFMATCH)]; ok {
			options[messages.IFMATCH] = val[0]
		}

		log.Debugf("Parsed URI, requestName=%+v, requestQuerys=%+v, options=%+v", requestName, requestQuerys, options)

		if requestName == "" || (!isClientConfigRequest(requestName) && !messages.IsRequest(requestName)) {
			fmt.Printf("dots_client.serverHandler -- %s is invalid request name \n", requestName)
			fmt.Printf("support messages: %s \n", messages.SupportRequest())
			errMessage := fmt.Sprintf("%s is invalid request name \n", requestName)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(errMessage))
			return
		}

		buff := new(bytes.Buffer)
		buff.ReadFrom(r.Body)

		var jsonData []byte = nil
		if 0 < buff.Len() {
			jsonData = buff.Bytes()
		}

		if r.Method == "POST" {
            switch requestName {
                case string(client_message.CLIENT_CONFIGURATION):
                    handleClientConfiguration(jsonData, env);
                case string(client_message.CLIENT_CONFIGURATION_HEARTBEAT):
                    handleClientConfigurationHeartBeat(jsonData, env);
                case string(client_message.CLIENT_CONFIGURATION_QBLOCK):
                    handleClientConfigurationQblock(jsonData, env);
                case string(client_message.CLIENT_CONFIGURATION_BLOCK):
                    handleClientConfigurationBlock(jsonData, env);
            }
			return
		}

		res, err := sendRequest(jsonData, requestName, r.Method, requestQuerys, env, options)
		if err != nil {
			fmt.Printf("dots_client.serverHandler -- %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}

		w.WriteHeader(res.StatusCode.HttpCode())
		w.Write(res.data)
	}
}

/**
* Check if request name is client config request
*/
func isClientConfigRequest(requestName string) bool {
	if requestName == string(client_message.CLIENT_CONFIGURATION) || requestName == string(client_message.CLIENT_CONFIGURATION_HEARTBEAT) ||
	   requestName == string(client_message.CLIENT_CONFIGURATION_QBLOCK) || requestName == string(client_message.CLIENT_CONFIGURATION_BLOCK) {
		return true
	}
	return false
}

/*
 * sendRequest is a function that sends requests to the server.
 */
func sendRequest(jsonData []byte, requestName, method string, queryParams []string, env *task.Env, options map[messages.Option]string) (res Response, err error) {
	if jsonData != nil {
		err = common.ValidateJson(requestName, string(jsonData))
		if err != nil {
			return
		}
	}
	code := messages.GetCode(requestName)
	libCoapType := messages.GetLibCoapType(requestName)

	var requestMessage RequestInterface
	switch messages.GetChannelType(requestName) {
	case messages.SIGNAL:
		requestMessage = NewRequest(code, libCoapType, method, requestName, queryParams, env, options)
	case messages.DATA:
		errorMsg := fmt.Sprintf("unsupported channel type error: %s", requestName)
		log.Errorf("dots_client.sendRequest -- %s", errorMsg)
		return res, errors.New(errorMsg)
	default:
		errorMsg := fmt.Sprintf("unknown channel type error: %s", requestName)
		log.Errorf("dots_client.sendRequest -- %s", errorMsg)
		return res, errors.New(errorMsg)
	}

	if jsonData != nil {
		err = requestMessage.LoadJson(jsonData)
		if err != nil {
			log.Errorf("dots_client.main -- JSON load error: %s", err.Error())
			return
		}
	}

	requestMessage.CreateRequest()
	log.Infof("dots_client.main -- request message: %+v", requestMessage)

	res = requestMessage.Send()
	return res, nil
}

var activeConWg sync.WaitGroup
var numberOfActive = 0

/*
 * connectionStateChange is a function to monitor the server connecion status.
 */
func connectionStateChange(_ net.Conn, connState http.ConnState) {
	if connState == http.StateActive {
		activeConWg.Add(1)
		numberOfActive += 1
	} else if connState == http.StateIdle || connState == http.StateHijacked {
		activeConWg.Done()
		numberOfActive -= 1
	}
	log.WithField("connection count", numberOfActive).Debug("receive http connection state event.")
}

func getDefaultCertPath(path string) string {
	packageRootPath := path + "/../"
	if goPath := os.Getenv("GOPATH"); goPath != "" {
		packageRootPath = goPath + "/src/github.com/nttdots/go-dots/"
	}

	log.WithField("root", packageRootPath).Debug("-- getDefaultCertPath")
	return packageRootPath + "certs/"
}


func restartConnection (env *task.Env) {
	log.Debug("Restart connection to server...")
	_,err := connectSignalChannel(env)
	if err != nil {
		log.WithError(err).Errorf("connectSignalChannel() failed")
		os.Exit(1)
	}
}

/*
 * Handle response from server with client environment
 * parameter:
 *  pdu   response pdu notification
 *  env   the client environment data
 */
func handleResponse(env *task.Env, pdu *libcoap.Pdu) {
	if env.IsPartialBlock() {
		log.Debugf("Unexpected incoming PDU: %s", pdu.ToString())
		env.SetIsPartialBlock(false)
		return
	}
    key := pdu.AsMapKey()
    t, ok := env.Requests()[key]
    if !ok {
		// If existed token, handle notification
		// Else handle forget notification
        if env.IsTokenExist(string(pdu.Token)) {
            handleNotification(env, nil, pdu)
        } else {
			observe, err := pdu.GetOptionIntegerValue(libcoap.OptionObserve)
			if err != nil {
				log.WithError(err).Warn("Failed to get observe option.")
				return
			}
			if observe >= 0 && pdu.Type == libcoap.TypeNon {
				log.Debug("Handle forget notification")
				env.CoapSession().HandleForgetNotification(pdu)
			} else {
				// Resource is deleted, then dots-client receive the response message from libcoap
				if pdu.Type == libcoap.TypeNon && pdu.Code == libcoap.ResponseNotFound && env.IsDeletedResource(string(pdu.Token)) {
					log.Debugf("Resource is deleted. Incoming PDU: %s", pdu.ToString())
					env.DeleteTokenOfDeletedResource(string(pdu.Token))
				} else {
					log.Debugf("Unexpected incoming PDU: %s", pdu.ToString())
				}
			}
        }
    } else if !t.IsStop() {
		delete(env.Requests(), key)
        t.Stop()
        t.GetResponseHandler()(t, pdu, env)
        // Reset current_missing_hb
	    env.SetCurrentMissingHb(0)
    }
}

/*
 * Handle request timeout with client environment
 * parameter:
 *  pdu   response pdu notification
 *  env   the client environment data
 */
func handleRequestTimeout(env *task.Env, sent *libcoap.Pdu) {
    key := sent.AsMapKey()
    t, ok := env.Requests()[key]

    if !ok {
        log.Info("Unexpected PDU: %v", sent)
    } else {
        t.Stop()
        log.Debugf("Session config request timeout")
        t.GetTimeoutHandler()(t, env)
    }
}

var config *dots_config.ClientSystemConfig
/**
* Load config file
*/
func loadConfig(env *task.Env) {
	env.SetMissingHbAllowed(config.DefaultSessionConfiguration.MissingHbAllowed)
	// Set max-retransmit, ack-timeout, ack-random-factor to libcoap
	env.SetRetransmitParams(config.DefaultSessionConfiguration.MaxRetransmit, decimal.NewFromFloat(config.DefaultSessionConfiguration.AckTimeout).Round(2),
		decimal.NewFromFloat(config.DefaultSessionConfiguration.AckRandomFactor).Round(2))
	env.SetIntervalBeforeMaxAge(config.IntervalBeforeMaxAge)
	if config.InitialRequestBlockSize != nil && *config.InitialRequestBlockSize >= 0 {
		env.SetInitialRequestBlockSize(config.InitialRequestBlockSize)
	}
	if config.SecondRequestBlockSize != nil && *config.SecondRequestBlockSize >= 0 {
		env.SetSecondRequestBlockSize(config.SecondRequestBlockSize)
	}
	// Set config of QBlock2 to libcoap
	if config.QBlockOption != nil {
		qblock := *config.QBlockOption
		env.SetRetransmitParamsForQBlock(qblock.QBlockSize, qblock.MaxPayloads, qblock.NonMaxRetransmit,
		qblock.NonTimeout, qblock.NonReceiveTimeout)
	}
}

func main() {

	log.Debug("parse arguments")
	flag.Parse()

	common.SetUpLogger()

	err := dots_config.LoadClientConfig(configFile)
	if err != nil {
		log.WithError(err).Errorf("LoadClientConfig() failed")
		os.Exit(1)
	}
	config = dots_config.GetSystemConfig()
	if config == nil {
		log.Warnf("The config is nil -> Stopped dots client")
		return
	}
	log.Debugf("dots client starting with config: %+v", config.String())

	serverIPs, err := net.LookupIP(server)
	if err != nil {
		log.Fatalf("Name Resolution failed: %s", server)
		os.Exit(1)
	}
	serverIP = serverIPs[0]

	if serverIP.To4() == nil {
		signalChannelAddress = fmt.Sprintf("[%s]:%d", server, signalChannelPort)
		dataChannelAddress = fmt.Sprintf("[%s]:%d", server, dataChannelPort)
	} else {
		signalChannelAddress = fmt.Sprintf("%s:%d", server, signalChannelPort)
		dataChannelAddress = fmt.Sprintf("%s:%d", server, dataChannelPort)
	}

	exists := func(filePath string) {
		_, err = os.Stat(filePath)
		if err != nil {
			log.Fatalf("dots_client.main --  file not found : %s", err.Error())
			os.Exit(1)
		}
	}

	if config.SecureFile != nil {
		certFile = config.SecureFile.CertFile
		clientCertFile = config.SecureFile.ClientCertFile
		clientKeyFile = config.SecureFile.ClientKeyFile
	}

	for _, filePath := range []string{certFile, clientCertFile, clientKeyFile} {
		exists(filePath)
	}

	env, err := connectSignalChannel(nil)
	if err != nil {
		log.WithError(err).Errorf("connectSignalChannel() failed")
		os.Exit(1)
	}

	log.Debugln("set http handler")

	http.HandleFunc("/server/", makeServerHandler(env))

	log.Infof("open unix domain socket on %s", socket)
	l, err := net.Listen("unix", socket)
	if err != nil {
		log.Errorf("dots_client.main -- socket listen error: %s", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	// Interruption handling
	stop := make(chan int, 1)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
	_:
		<-c
		activeConWg.Wait()
		if err := l.Close(); err != nil {
			log.Errorf("error: %v", err)
			os.Exit(1)
		}
		stop <- 0
	}()

	srv := &http.Server{Handler: nil, ConnState: connectionStateChange}
	go srv.Serve(l)

	// Run restful api server to service external systems
	address := config.ClientRestfulApiConfiguration.RestfulApiAddress + config.ClientRestfulApiConfiguration.RestfulApiPort
	go restful_router.ListenRestfulApi(address, makeServerHandler(env))

	// Load session configuration
	loadConfig(env)
	env.Run(task.NewHeartBeatTask(
		time.Duration(config.DefaultSessionConfiguration.HeartbeatInterval) * time.Second,
		config.DefaultSessionConfiguration.MissingHbAllowed,
		time.Duration(config.DefaultSessionConfiguration.AckTimeout) * time.Second,
		heartbeatResponseHandler,
		heartbeatTimeoutHandler))
loop:
	for {
		select {
		case e := <- env.EventChannel():
			e.Handle(env)
		case <- stop:
			break loop
		default:
			env.CoapContext().RunOnce(time.Duration(100) * time.Millisecond)
			CheckReplacingSession(env)
		}
	}
	cleanupSignalChannel(env.CoapContext(), env.CoapSession())
}

func CheckReplacingSession(env *task.Env) {
	isReplace := env.CheckSessionReplacement()
	if isReplace {
        loadConfig(env)
		env.Run(task.NewHeartBeatTask(
				time.Duration(config.DefaultSessionConfiguration.HeartbeatInterval) * time.Second,
				config.DefaultSessionConfiguration.MissingHbAllowed,
				time.Duration(config.DefaultSessionConfiguration.AckTimeout) * time.Second,
				heartbeatResponseHandler,
				heartbeatTimeoutHandler))
	}
}

// Handle set config mode by request from client controller (idle or mitigating)
func handleClientConfiguration(jsonData []byte, env *task.Env) {
	var clientConfig *client_message.ClientConfigRequest
	err := json.Unmarshal(jsonData, &clientConfig)
	if err != nil {
		log.Errorf("Failed to parse json data : %+v", err)
		return
	}
	mode := clientConfig.SessionConfig.Mode
	log.Debugf("Session config mode: %s", mode)
	if mode == string(client_message.IDLE) || mode == string(client_message.MITIGATING) {
		log.Debugf("Session config mode is valid. Switch to new session config mode: %s", mode)
		env.SetSessionConfigMode(mode)
	} else {
		log.Debug("Session config mode is invalid")
	}
}

// Handle set heart beat parameters by request from client controller
func handleClientConfigurationHeartBeat(jsonData []byte, env *task.Env) {
    var clientConfigHeartBeat *client_message.ClientConfigHeartBeatRequest
	err := json.Unmarshal(jsonData, &clientConfigHeartBeat)
	if err != nil {
		log.Errorf("Failed to parse json data : %+v", err)
		return
	}
	heartbeatInterval := clientConfigHeartBeat.SessionConfigHeartBeat.HeartBeatInterval
	missingHbAllowed := clientConfigHeartBeat.SessionConfigHeartBeat.MissingHbAllowed
	maxRetransmit := clientConfigHeartBeat.SessionConfigHeartBeat.MaxRetransmit
	ackTimeout := decimal.NewFromFloat(clientConfigHeartBeat.SessionConfigHeartBeat.AckTimeout).Round(2)
	ackRandomFactor := decimal.NewFromFloat(clientConfigHeartBeat.SessionConfigHeartBeat.AckRandomFactor).Round(2)
	log.Debugf("Set new parameter for heart beat. Restart ping task with: \n%s", clientConfigHeartBeat.SessionConfigHeartBeat.String())
	// Set max-retransmit, ack-timeout, ack-random-factor to libcoap
	env.SetRetransmitParams(maxRetransmit, ackTimeout, ackRandomFactor)
	pingTimeout, _ := ackTimeout.Float64()
	env.StopHeartBeat()
	env.SetMissingHbAllowed(missingHbAllowed)
	env.Run(task.NewHeartBeatTask(
			time.Duration(heartbeatInterval)* time.Second,
			missingHbAllowed,
			time.Duration(pingTimeout) * time.Second,
			heartbeatResponseHandler,
			heartbeatTimeoutHandler))
}

// Handle set QBlock paramters by request from client controller
func handleClientConfigurationQblock(jsonData []byte, env *task.Env) {
	var clientConfigQBlock *client_message.ClientConfigQBlockRequest
	err := json.Unmarshal(jsonData, &clientConfigQBlock)
	if err != nil {
		log.Errorf("Failed to parse json data : %+v", err)
		return
	}
	qblock := clientConfigQBlock.SessionConfigQBlock
	// Validate qblock size
	if qblock.QBlockSize < 0 || qblock.QBlockSize > 7 {
		log.Errorf("q-block-size: %+v is invalid", qblock.QBlockSize)
		return
	}
	log.Debugf("Set new parameter for qblock option: \n%s", qblock.String())
	env.SetRetransmitParamsForQBlock(qblock.QBlockSize, qblock.MaxPayload, qblock.NonMaxRetransmit, 
		qblock.NonTimeout, qblock.NonReceiveTimeout)
	env.SetInitialRequestBlockSize(nil)
}

// Handle set Block paramter by request from client controller
func handleClientConfigurationBlock(jsonData []byte, env *task.Env) {
	var clientConfigBlock *client_message.ClientConfigBlockRequest
	err := json.Unmarshal(jsonData, &clientConfigBlock)
	if err != nil {
		log.Errorf("Failed to parse json data : %+v", err)
		return
	}
	blockSize := clientConfigBlock.SessionConfigBlock.BlockSize
	// Validate block size
	if blockSize < 0 || blockSize > 7 {
		log.Errorf("block-size: %+v is invalid", blockSize)
		return
	}
	log.Debugf("Set new parameter for block option with block-size: %d", blockSize)
	env.SetInitialRequestBlockSize(&blockSize)
}