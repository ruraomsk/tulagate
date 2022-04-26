package device

import (
	"encoding/json"
	"time"

	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/tulagate/controller"
	"github.com/ruraomsk/tulagate/tulastat"
)

func (n *nowStat) init() {
	n.stat = make([]int, 16)
}
func (n *nowStat) getMGRword() int {
	res := 0
	for i := 15; i >= 0; i-- {
		if n.stat[i] != 0 {
			res |= 1
			res = res << 1
		}
	}
	return res
}
func (d *Device) initStatistic() {
	d.Stat.nowStat.init()
	d.Stat.finStat.init()
	d.Stat.count = 0
}
func (d *Device) addStat(st []int) {
	for i := 0; i < len(st); i++ {
		d.Stat.nowStat.stat[i] = st[i]
		d.Stat.finStat.stat[i] = d.Stat.finStat.stat[i] + st[i]
	}
	d.Stat.count++
}
func (d *Device) sendStatistics(ptime int) {
	s := pudge.Statistic{Period: ptime / d.Stat.interval, Type: d.Stat.tp, TLen: d.Stat.interval / 60,
		Hour: ptime / 3600, Min: (ptime % 3600) / 60, Datas: make([]pudge.DataStat, 0)}
	// logger.Debug.Print(d.Region, d.Stat.finStat)
	good := 1
	if d.Stat.count > 45 {
		good = 0
	}
	for i, v := range d.Stat.finStat.stat {
		if d.Stat.tp == 2 {
			s.Datas = append(s.Datas, pudge.DataStat{Chanel: i + 1, Status: good, Speed: v})
		} else {
			s.Datas = append(s.Datas, pudge.DataStat{Chanel: i + 1, Status: good, Intensiv: v})
		}
	}
	// logger.Debug.Print(s)
	tulastat.StatisticChan <- tulastat.RecordStat{Region: d.Region, Stat: s}
	d.initStatistic()
}
func (d *Device) executeAddStat(message controller.MessageFromAmi) {
	var setter controller.ChanelStat
	d.LastReciveStat = time.Now()
	json.Unmarshal([]byte(message.Body), &setter)
	d.addStat(setter.Chanels[:])
}
func TimeNowOfSecond() int {
	return time.Now().Hour()*3600 + time.Now().Minute()*60 + time.Now().Second()
}
func TimeToSeconds(p time.Time) int {
	return p.Hour()*3600 + p.Minute()*60 + p.Second()
}
