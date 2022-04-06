package db

import (
	"database/sql"
	"fmt"
	"sync"

	//Инициализатор постргресса
	_ "github.com/lib/pq"
	"github.com/ruraomsk/ag-server/comm"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/tulagate/controller"
)

var db *sql.DB
var mutex sync.Mutex
var crosses map[pudge.Region]pudge.Cross
var ctrls map[int]pudge.Controller

func GetChanelForMessage(extid string) (chan controller.MessageFromAmi, error) {
	return nil, fmt.Errorf("нет такого устройства %s", extid)
}
func GetChanReceivePhases(id int) (chan comm.DevPhases, error) {
	return nil, fmt.Errorf("нет такого устройства %d", id)
}
func Starter(dkset *controller.DKSet) error {

	return nil
}
