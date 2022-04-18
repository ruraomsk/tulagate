package db

import (
	"encoding/json"

	"github.com/ruraomsk/ag-server/binding"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
)

var create = `
CREATE TABLE IF NOT EXISTS public."base"
(
    region integer NOT NULL,
    area integer NOT NULL,
    id integer NOT NULL,
    state jsonb NOT NULL
)
`

func MoveData(cross *pudge.Cross, baseCross *pudge.Cross) {
	cross.Arrays.SetDK = baseCross.Arrays.SetDK
	cross.Arrays.DaySets = baseCross.Arrays.DaySets
	cross.Arrays.WeekSets = baseCross.Arrays.WeekSets
	cross.Arrays.MonthSets = baseCross.Arrays.MonthSets
}
func touchBase() {
	_, err := dbBase.Query("select state from public.\"base\" limit 1;")
	if err != nil {
		dbBase.Exec(create)
	}
}

func SetBasePlan(region pudge.Region, dk binding.SetDK, pk int) {
	baseMutex.Lock()
	defer baseMutex.Unlock()
	rows, err := dbBase.Query("select state from public.\"base\" where region=$1 and area=$2 and id=$3;", region.Region, region.Area, region.ID)
	if err != nil {
		logger.Error.Print(err.Error())
		return
	}
	var cross pudge.Cross
	var state []byte
	found := false
	for rows.Next() {
		found = true
		rows.Scan(&state)
		err = json.Unmarshal(state, &cross)
		if err != nil {
			logger.Error.Print(err.Error())
			return
		}
	}
	rows.Close()
	if !found {
		//Значит произведем заготовку
		cross, _ = GetCross(region)
		ClearPKs(&cross)
	}
	for i := 0; i < len(dk.DK); i++ {
		if dk.DK[i].Pk == pk {
			for j := 0; j < len(cross.Arrays.SetDK.DK); j++ {
				if cross.Arrays.SetDK.DK[j].Pk == pk {
					cross.Arrays.SetDK.DK[j] = dk.DK[i]
					break
				}
			}
			break
		}
	}
	state, _ = json.Marshal(cross)
	if found {
		db.Exec("update public.\"base\" set state=$1 where region=$2 and area=$3 and id=$4;",
			string(state), region.Region, region.Area, region.ID)
	} else {
		db.Exec("insert into public.\"base\" (region, area,  id, state) values ($1,$2,$3,$4);",
			region.Region, region.Area, region.ID, string(state))
	}
}
func GetStartCross(region pudge.Region) (pudge.Cross, error) {
	baseMutex.Lock()
	defer baseMutex.Unlock()
	rows, err := dbBase.Query("select state from public.\"base\" where region=$1 and area=$2 and id=$3;", region.Region, region.Area, region.ID)
	if err != nil {
		logger.Error.Print(err.Error())
		return pudge.Cross{}, err
	}
	var cross pudge.Cross
	var state []byte
	found := false
	for rows.Next() {
		found = true
		rows.Scan(&state)
		err = json.Unmarshal(state, &cross)
		if err != nil {
			logger.Error.Print(err.Error())
			return pudge.Cross{}, err
		}
	}
	rows.Close()
	if !found {
		//Значит произведем заготовку
		cross, _ = GetCross(region)
		ClearPKs(&cross)
	}
	if !found {
		state, _ = json.Marshal(cross)
		db.Exec("insert into public.\"base\" (region, area,  id, state) values ($1,$2,$3,$4);",
			region.Region, region.Area, region.ID, string(state))
	}
	return cross, nil
}
func ClearPKs(cross *pudge.Cross) {
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
}
