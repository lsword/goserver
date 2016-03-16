package httpserver

import (
	//"encoding/json"
	"fmt"
	"io"
	//"io/ioutil"
	"net"
	"net/http"
	_ "net/http/pprof"

	//"github.com/golang/net/netutil"
	"github.com/gorilla/mux"
	"goserver/config"
	"goserver/dbserver"
	"goserver/log"
	"goserver/pkg/version"
)

type HttpApiFunc func(version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error

func createRouter() (*mux.Router, error) {
	r := mux.NewRouter()
	m := map[string]map[string]HttpApiFunc{
		"GET": {
			"/serverinfo":   getServerInfo,
			"/serverconfig": getServerConfig,
			"/testquery":    getTestQuery,
			"/testexec":     getTestExec,
			//"/container/{containerid:.*}/json": getContainerById,
		},
		"POST": {
		//"/containers/create":  postContainersCreate,
		//"/containers/start":   postContainersStart,
		//"/containers/stop":    postContainersStop,
		//"/containers/pause":   postContainersPause,
		//"/containers/unpause": postContainersUnpause,
		//"/containers/delete":  postContainersDelete,
		},
		"OPTIONS": {
			"": optionsHandler,
		},
	}

	for method, routes := range m {
		for route, fct := range routes {
			log.Debugf("Registering %s, %s", method, route)

			localRoute := route
			localFct := fct
			localMethod := method

			f := makeHttpHandler(true, localMethod, localRoute, localFct, true, version.Version(""))

			if localRoute == "" {
				r.Methods(localMethod).HandlerFunc(f)
			} else {
				log.Infof("localRoute:%s, localMethod:%s", localRoute, localMethod)
				r.Path("/v{version:[0-9.]+}" + localRoute).Methods(localMethod).HandlerFunc(f)
				r.Path(localRoute).Methods(localMethod).HandlerFunc(f)
			}
		}
	}

	return r, nil
}

func makeHttpHandler(logging bool, localMethod string, localRoute string, handlerFunc HttpApiFunc, enableCors bool, dockerCenterVersion version.Version) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if logging {
			log.Debugf("%s %s", r.Method, r.RequestURI)
		}

		if enableCors {
			writeCorsHeaders(w, r)
		}

		if err := handlerFunc("", w, r, mux.Vars(r)); err != nil {
			log.Errorf("Error: %s", err)
		}
	}
}

func optionsHandler(version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	w.WriteHeader(http.StatusOK)
	return nil
}

func writeCorsHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.Header().Add("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS")
}

func ping(version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	_, err := w.Write([]byte{'O', 'K'})
	return err
}

func Run(c *config.Config) {
	if c.HttpServer.Switch != "on" {
		return
	}

	r, err := createRouter()
	if err != nil {
		log.Fatal("create router error")
		return
	}

	httpSrv := http.Server{
		Addr:    fmt.Sprintf(":%d", c.HttpServer.Port),
		Handler: r,
	}

	l, _ := net.Listen("tcp", fmt.Sprintf(":%d", c.HttpServer.Port))
	//l = LimitListener(l, 100)
	go func() {
		err = httpSrv.Serve(l)
		if err != nil {
			log.Fatal("http serve error")
			return
		}
	}()
}

func getServerInfo(version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if r.Method != "GET" {
		return fmt.Errorf("Invalid Method")
	}

	log.Info("receive get server info request")

	io.WriteString(w, dbserver.Status())
	return nil
}

func getServerConfig(version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if r.Method != "GET" {
		return fmt.Errorf("Invalid Method")
	}

	io.WriteString(w, fmt.Sprintf("%s", config.ConfigJson()))
	return nil
}

func getTestQuery(version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if r.Method != "GET" {
		return fmt.Errorf("Invalid Method")
	}

	db := dbserver.GetDatabase()
	_, rows, _ := db.QueryData("mysql1", "select * from test where id=5")
	result := ""
	for row := range *rows {
		for k, v := range (*rows)[row] {
			if v == nil {
				result = result + k + ":NULL" + "; "
				continue
			}
			switch k {
			case "id":
				result = result + fmt.Sprintf("%s:%s;", k, string(v.([]byte)))
				break
			case "name":
				result = result + fmt.Sprintf("%s:%s;", k, string(v.([]byte)))
				break
			}
		}
		result = result + "\n"
	}
	io.WriteString(w, result)

	return nil
}

func getTestExec(version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if r.Method != "GET" {
		return fmt.Errorf("Invalid Method")
	}

	log.Info("receive get server info request")

	db := dbserver.GetDatabase()
	lastid, affectrow, err := db.Exec("mysql1", "insert into test(id,name) values (?,?)", 3, "123")
	if err != nil {
		io.WriteString(w, fmt.Sprintf("lastid:%d, affectrow:%d, error:%s", lastid, affectrow, err.Error()))
	} else {
		io.WriteString(w, fmt.Sprintf("lastid:%d, affectrow:%d", lastid, affectrow))
	}

	return nil
}

/*
func getDockerAgentList(version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if r.Method != "GET" {
		return fmt.Errorf("Invalid Method")
	}

	io.WriteString(w, string(data.GetDockerAgentMap().Json()))

	return nil
}

func getContainerById(version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	var res protocols.HttpQueryContainerRsp
	res.Result = -1

	if r.Method != "GET" {
		res.Info = "Invalid Method"
		resBody, _ := json.Marshal(res)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Info)
	}

	containerid := vars["containerid"]
	if containerid == "" {
		res.Info = "Invalid Argument"
		resBody, _ := json.Marshal(res)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Info)
	}

	//根据容器id查找其所在的dockeragent
	da := data.FindDockerAgentByContainerId(containerid)
	if da == nil {
		log.Errorf("getContainerById data.FindDockerAgentByContainerId error:can not find dockeragent by containerid:%s", containerid)
		res.Info = "Start container task failed"
		resBody, _ := json.Marshal(res)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Info)
	}

	containerinfo, err := natsclient.InspectContainer(da.Ip, containerid)
	if err != nil {
		res.Info = fmt.Sprintf("InspectContainer err:%s", err)
		resBody, _ := json.Marshal(res)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Info)
	}

	res.Result = 0
	res.Info = ""
	res.AgentIp = da.Ip
	res.ContainerInfo = containerinfo
	resBody, _ := json.MarshalIndent(res, "", "  ")
	io.WriteString(w, string(resBody))
	return nil

}

func postContainersCreate(version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	ress := make([]protocols.HttpTaskRsp, 0)

	if r.Method != "POST" {
		res := protocols.GenerateTaskRsp(-2, "", "", "Invalid Method")
		ress = append(ress, res)
		resBody, _ := json.Marshal(ress)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Error)
	}

	msgbody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("postContainersCreate: read post body error: ", err)
		res := protocols.GenerateTaskRsp(-2, "", "", "Invalid Arguments")
		ress = append(ress, res)
		resBody, _ := json.Marshal(ress)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Error)
	}

	log.Debugf("postContainersCreate msgbody:%s", msgbody)

	fmt.Println(string(msgbody))
	var hccr protocols.HttpCreateContainerReq
	err = json.Unmarshal(msgbody, &hccr)
	if err != nil {
		log.Errorf("postContainersCreate unmarshal error:%s", err)
		res := protocols.GenerateTaskRsp(-2, "", "", "Invalid Arguments")
		ress = append(ress, res)
		resBody, _ := json.Marshal(ress)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Error)
	}

	//创建并分配子任务
	resource := &data.Resource{
		Cpu:  hccr.CpuLimit,
		Mem:  hccr.MemLimit,
		Disk: hccr.DiskLimit,
	}
	var forbiddenIps []string
	for _, fip := range hccr.ForbiddenIps {
		forbiddenIps = append(forbiddenIps, fip)
	}

	for i := 0; i < hccr.Count; i++ {
		da := data.GetDockerAgentMap().FindDockerAgent(resource, &forbiddenIps)
		if da == nil {
			continue
		} else {
			//通知dockeragent创建容器
			taskRes := natsclient.Request(da.Ip, protocols.Nats_AgentTaskType_CreateContainer, string(msgbody))
			forbiddenIps = append(forbiddenIps, da.Ip)
			ress = append(ress, taskRes)
		}
	}

	resBody, _ := json.Marshal(ress)
	log.Debugf("Create container output: %s", string(resBody))
	io.WriteString(w, string(resBody))
	return nil
}

func postContainersStart(version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if r.Method != "POST" {
		res := protocols.GenerateTaskRsp(-2, "", "", "Invalid Method")
		resBody, _ := json.Marshal(res)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Error)
	}

	msgbody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("postContainersStart: read post body error: ", err)
		res := protocols.GenerateTaskRsp(-2, "", "", "Invalid Arguments")
		resBody, _ := json.Marshal(res)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Error)
	}

	log.Debugf("postContainersStart msgbody:%s", msgbody)

	var hscr protocols.HttpStartContainerReq
	err = json.Unmarshal([]byte(msgbody), &hscr)
	if err != nil {
		log.Errorf("postContainersStart unmarshal error:%s", err)
		res := protocols.GenerateTaskRsp(-2, "", "", "Invalid Arguments")
		resBody, _ := json.Marshal(res)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Error)
	}

	//根据容器id查找其所在的dockeragent
	da := data.FindDockerAgentByContainerId(hscr.ContainerId)
	if da == nil {
		log.Errorf("postContainersStart data.NewStartContainerTask error:can not find dockeragent by containerid:%s", hscr.ContainerId)
		res := protocols.GenerateTaskRsp(-2, "", "", "can not find dockeragent by containerid:"+hscr.ContainerId)
		resBody, _ := json.Marshal(res)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Error)
	}

	response := natsclient.Request(da.Ip, protocols.Nats_AgentTaskType_StartContainer, string(msgbody))
	resBody, _ := json.Marshal(response)
	io.WriteString(w, string(resBody))
	return nil
}

func postContainersStop(version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if r.Method != "POST" {
		res := protocols.GenerateTaskRsp(-2, "", "", "Invalid Method")
		resBody, _ := json.Marshal(res)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Error)
	}

	msgbody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("postContainersStop: read post body error: ", err)
		res := protocols.GenerateTaskRsp(-2, "", "", "Invalid Arguments")
		resBody, _ := json.Marshal(res)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Error)
	}

	log.Debugf("postContainersStop msgbody:%s", msgbody)

	var hscr protocols.HttpStopContainerReq
	err = json.Unmarshal([]byte(msgbody), &hscr)
	if err != nil {
		log.Errorf("postContainersStop unmarshal error:%s", err)
		res := protocols.GenerateTaskRsp(-2, "", "", "Invalid Arguments")
		resBody, _ := json.Marshal(res)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Error)
	}

	//根据容器id查找其所在的dockeragent
	da := data.FindDockerAgentByContainerId(hscr.ContainerId)
	if da == nil {
		log.Errorf("postContainersStart data.NewStartContainerTask error:can not find dockeragent by containerid:%s", hscr.ContainerId)
		res := protocols.GenerateTaskRsp(-2, "", "", "can not find dockeragent by containerid:"+hscr.ContainerId)
		resBody, _ := json.Marshal(res)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Error)
	}

	response := natsclient.Request(da.Ip, protocols.Nats_AgentTaskType_StopContainer, string(msgbody))
	resBody, _ := json.Marshal(response)
	io.WriteString(w, string(resBody))
	return nil
}

func postContainersPause(version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if r.Method != "POST" {
		res := protocols.GenerateTaskRsp(-2, "", "", "Invalid Method")
		resBody, _ := json.Marshal(res)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Error)
	}

	msgbody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("postContainersPause: read post body error: ", err)
		res := protocols.GenerateTaskRsp(-2, "", "", "Invalid Arguments")
		resBody, _ := json.Marshal(res)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Error)
	}

	log.Debugf("postContainersPause msgbody:%s", msgbody)

	var hscr protocols.HttpPauseContainerReq
	err = json.Unmarshal([]byte(msgbody), &hscr)
	if err != nil {
		log.Errorf("postContainersPause unmarshal error:%s", err)
		res := protocols.GenerateTaskRsp(-2, "", "", "Invalid Arguments")
		resBody, _ := json.Marshal(res)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Error)
	}

	//根据容器id查找其所在的dockeragent
	da := data.FindDockerAgentByContainerId(hscr.ContainerId)
	if da == nil {
		log.Errorf("postContainersPause data.data.FindDockerAgentByContainerId error:can not find dockeragent by containerid:%s", hscr.ContainerId)
		res := protocols.GenerateTaskRsp(-2, "", "", "can not find dockeragent by containerid:"+hscr.ContainerId)
		resBody, _ := json.Marshal(res)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Error)
	}

	response := natsclient.Request(da.Ip, protocols.Nats_AgentTaskType_PauseContainer, string(msgbody))
	resBody, _ := json.Marshal(response)
	io.WriteString(w, string(resBody))
	return nil
}

func postContainersUnpause(version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if r.Method != "POST" {
		res := protocols.GenerateTaskRsp(-2, "", "", "Invalid Method")
		resBody, _ := json.Marshal(res)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Error)
	}

	msgbody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("postContainersPause: read post body error: ", err)
		res := protocols.GenerateTaskRsp(-2, "", "", "Invalid Arguments")
		resBody, _ := json.Marshal(res)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Error)
	}

	log.Debugf("postContainersPause msgbody:%s", msgbody)

	var hscr protocols.HttpUnpauseContainerReq
	err = json.Unmarshal([]byte(msgbody), &hscr)
	if err != nil {
		log.Errorf("postContainersPause unmarshal error:%s", err)
		res := protocols.GenerateTaskRsp(-2, "", "", "Invalid Arguments")
		resBody, _ := json.Marshal(res)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Error)
	}

	//根据容器id查找其所在的dockeragent
	da := data.FindDockerAgentByContainerId(hscr.ContainerId)
	if da == nil {
		log.Errorf("postContainersPause data.data.FindDockerAgentByContainerId error:can not find dockeragent by containerid:%s", hscr.ContainerId)
		res := protocols.GenerateTaskRsp(-2, "", "", "can not find dockeragent by containerid:"+hscr.ContainerId)
		resBody, _ := json.Marshal(res)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Error)
	}

	response := natsclient.Request(da.Ip, protocols.Nats_AgentTaskType_UnpauseContainer, string(msgbody))
	resBody, _ := json.Marshal(response)
	io.WriteString(w, string(resBody))

	return nil
}

func postContainersDelete(version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if r.Method != "POST" {
		res := protocols.GenerateTaskRsp(-2, "", "", "Invalid Method")
		resBody, _ := json.Marshal(res)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Error)
	}

	msgbody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("postContainersPause: read post body error: ", err)
		res := protocols.GenerateTaskRsp(-2, "", "", "Invalid Arguments")
		resBody, _ := json.Marshal(res)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Error)
	}

	log.Debugf("postContainersDelete msgbody:%s", string(msgbody))
	var hdcr protocols.HttpDeleteContainerReq
	err = json.Unmarshal(msgbody, &hdcr)
	if err != nil {
		log.Errorf("postContainersPause unmarshal error:%s", err)
		res := protocols.GenerateTaskRsp(-2, "", "", "Invalid Arguments")
		resBody, _ := json.Marshal(res)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Error)
	}

	//根据容器id查找其所在的dockeragent
	da := data.FindDockerAgentByContainerId(hdcr.ContainerId)
	if da == nil {
		log.Errorf("deleteContainers data.NewDeleteContainerTask error:can not find dockeragent by containerid:%s", hdcr.ContainerId)
		res := protocols.GenerateTaskRsp(-2, "", "", "can not find dockeragent by containerid:"+hdcr.ContainerId)
		resBody, _ := json.Marshal(res)
		io.WriteString(w, string(resBody))
		return fmt.Errorf(res.Error)
	}

	response := natsclient.Request(da.Ip, protocols.Nats_AgentTaskType_DeleteContainer, string(msgbody))
	resBody, _ := json.Marshal(response)
	io.WriteString(w, string(resBody))
	return nil
}
*/
