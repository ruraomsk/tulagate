package creator

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ruraomsk/ag-server/binding"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/tulagate/controller"
	"github.com/ruraomsk/tulagate/setup"
)

var message = `Вы начинаете подготовку к работе системы TulaGate
Во время этой процедуры будут приведены в начальное состояние перекрестки и устройства
а именно будут утеряны все назначенные планы координации
Кроме этого Вам необходимо остановить программы ag-server и TLServer.
Если вы не этого хотели введите Нет, иначе наберите Да и продолжим....
`
var deleteTableBase = `
DROP TABLE IF EXISTS public."base";
`

func Creator(dks *controller.DKSet) {
	var repl string
	var db *sql.DB
	var err error
	fmt.Println(message)
	fmt.Scan(&repl)
	if strings.Compare(strings.ToUpper(repl), "ДА") != 0 {
		fmt.Println("Работа прервана...")
		os.Exit(-1)
	}
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
		break
	}
	fmt.Println("Соединение с базой данных установлено...")
	db.Exec(deleteTableBase)
	fmt.Println("Удалена таблица базовых привязок...")
	for _, v := range dks.DKSets {
		rows, err := db.Query("select state from public.\"cross\" where region=$1 and area=$2 and id=$3;", setup.Set.Region, v.Area, v.ID)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(-1)
		}
		var state []byte
		var cross pudge.Cross
		found := false
		for rows.Next() {
			rows.Scan(&state)
			err = json.Unmarshal(state, &cross)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(-1)
			}
			cross.Arrays.MonthSets = *binding.NewYearSets()
			for i := 0; i < len(cross.Arrays.MonthSets.MonthSets); i++ {
				for j := 0; j < len(cross.Arrays.MonthSets.MonthSets[i].Days); j++ {
					cross.Arrays.MonthSets.MonthSets[i].Days[j] = 1
				}
			}
			cross.Arrays.WeekSets = *binding.NewWeekSets()
			for i := 0; i < len(cross.Arrays.WeekSets.WeekSets); i++ {
				for j := 0; j < len(cross.Arrays.WeekSets.WeekSets[i].Days); j++ {
					cross.Arrays.WeekSets.WeekSets[i].Days[j] = 1
				}
			}
			cross.Arrays.DaySets = *binding.NewDaySet()
			cross.Arrays.DaySets.DaySets[0].Count = 1
			cross.Arrays.DaySets.DaySets[0].Number = 1
			cross.Arrays.DaySets.DaySets[0].Lines[0].Hour = 24
			cross.Arrays.DaySets.DaySets[0].Lines[0].Min = 0
			cross.Arrays.DaySets.DaySets[0].Lines[0].PKNom = 1

			cross.Arrays.SetDK = *binding.NewSetDK()
			found = true
		}
		if !found {
			fmt.Printf("Нет какого перекрестка %d %d %d\n", setup.Set.Region, v.Area, v.ID)
			os.Exit(-1)
		}
		rows.Close()
		buffer, err := json.Marshal(&cross)
		if err != nil {
			fmt.Printf("%d %d %d %s\n", setup.Set.Region, v.Area, v.ID, err.Error())
			os.Exit(-1)
		}
		_, err = db.Exec("update public.\"cross\" set state=$1 where region=$2 and area=$3 and id=$4;", string(buffer),
			setup.Set.Region, v.Area, v.ID)
		if err != nil {
			fmt.Printf("%d %d %d %s", setup.Set.Region, v.Area, v.ID, err.Error())
			os.Exit(-1)
		}
		fmt.Printf("Перекресток %d %d %d обработан\n", setup.Set.Region, v.Area, v.ID)
	}
}
