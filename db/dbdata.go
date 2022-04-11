package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	//Инициализатор постргресса
	_ "github.com/lib/pq"

	"github.com/ruraomsk/ag-server/comm"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/tulagate/controller"
	"github.com/ruraomsk/tulagate/setup"
)

var (
	db           *sql.DB = nil
	mutex        sync.Mutex
	temp_crosses map[pudge.Region]pudge.Cross
	temp_ctrls   map[int]pudge.Controller
	crosses      map[pudge.Region]pudge.Cross
	ctrls        map[int]pudge.Controller
	status       map[int]vstatus
	err          error
	dkset        *controller.DKSet
	phaseschans  map[int]chan comm.DevPhases
	messagechans map[string]chan controller.MessageFromAmi
)

type vstatus struct {
	id          int
	description string
	control     bool
}

func GetChanelForMessage(extid string) (chan controller.MessageFromAmi, error) {
	r, is := messagechans[extid]
	if !is {
		return nil, fmt.Errorf("нет такого устройства %s", extid)
	}
	return r, nil
}
func AddChanelForMessage(extid string, ch chan controller.MessageFromAmi) {
	messagechans[extid] = ch
}

func GetChanReceivePhases(id int) (chan comm.DevPhases, error) {
	r, is := phaseschans[id]
	if !is {
		return nil, fmt.Errorf("нет такого устройства %d", id)
	}
	return r, nil
}
func AddChanReceivePhases(id int, ch chan comm.DevPhases) {
	phaseschans[id] = ch
}
func GetControlStatus(id int) bool {
	return status[id].control
}
func GetDescription(id int) string {
	return status[id].description
}
func GetCross(region pudge.Region) (pudge.Cross, error) {
	mutex.Lock()
	defer mutex.Unlock()
	c, is := crosses[region]
	if !is {
		return pudge.Cross{}, fmt.Errorf("нет %v", region)
	}
	return c, nil
}

func GetController(id int) (pudge.Controller, error) {
	mutex.Lock()
	defer mutex.Unlock()
	c, is := ctrls[id]
	if !is {
		return pudge.Controller{ID: id, StatusConnection: false, LastOperation: time.Unix(0, 0)}, fmt.Errorf("нет контроллера %d", id)
	}
	return c, nil
}

func Starter(dks *controller.DKSet, next chan interface{}) {
	dkset = dks
	temp_crosses = make(map[pudge.Region]pudge.Cross)
	temp_ctrls = make(map[int]pudge.Controller)
	status = make(map[int]vstatus)
	phaseschans = make(map[int]chan comm.DevPhases)
	messagechans = make(map[string]chan controller.MessageFromAmi)
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)
	for {
		db, err = sql.Open("postgres", dbinfo)
		if err != nil {
			logger.Error.Print(err.Error())
			time.Sleep(10 * time.Second)
			continue
		}
		err = db.Ping()
		if err != nil {
			logger.Error.Print(err.Error())
			time.Sleep(10 * time.Second)
			continue
		}
		rows, err := db.Query("select id,description,control from public.status;")
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(-1)
		}
		for rows.Next() {
			var id int
			var description string
			var control bool
			rows.Scan(&id, &description, &control)
			status[id] = vstatus{id: id, description: description, control: control}
		}
		rows.Close()
		loadCrossAndCtrl()
		if len(crosses) == 0 {
			fmt.Println("нет данных по массиву настройки")
			os.Exit(-1)
		}
		next <- 1
		ticker := time.NewTicker(time.Duration(setup.Set.DataBase.Step) * time.Second)
		for {
			<-ticker.C
			loadCrossAndCtrl()
		}

	}
}
func loadCrossAndCtrl() {
	// logger.Info.Print("load All DB")
	for _, v := range dkset.DKSets {
		w := fmt.Sprintf("select state from public.\"cross\" where region=%d and area=%d and id=%d;", setup.Set.Region, v.Area, v.ID)
		rows, err := db.Query(w)
		if err != nil {
			logger.Error.Printf("%s %s", w, err.Error())
			continue
		}
		for rows.Next() {
			var cross pudge.Cross
			var buff []byte
			rows.Scan(&buff)
			err = json.Unmarshal(buff, &cross)
			if err != nil {
				logger.Error.Print(err.Error())
				continue
			}
			region := pudge.Region{Region: setup.Set.Region, Area: v.Area, ID: v.ID}
			// logger.Info.Print(region)
			temp_crosses[region] = cross

			dev, err := db.Query(fmt.Sprintf("select device from public.devices where id=%d;", cross.IDevice))
			if err != nil {
				logger.Error.Print(err.Error())
				continue
			}
			for dev.Next() {
				var buf []byte
				var ctrl pudge.Controller
				dev.Scan(&buf)
				// logger.Debug.Print(string(buf))
				err = json.Unmarshal(buf, &ctrl)
				if err != nil {
					logger.Error.Printf("%d %s", cross.IDevice, err.Error())
					continue
				}
				temp_ctrls[cross.IDevice] = ctrl
			}
			dev.Close()
		}
		rows.Close()
	}

	mutex.Lock()
	crosses = make(map[pudge.Region]pudge.Cross)
	ctrls = make(map[int]pudge.Controller)
	for k, v := range temp_crosses {
		crosses[k] = v
	}
	for k, v := range temp_ctrls {
		ctrls[k] = v
	}
	mutex.Unlock()

}
