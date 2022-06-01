package controller

import (
	"fmt"

	"github.com/ruraomsk/ag-server/binding"
	"github.com/ruraomsk/ag-server/logger"
)

func (c *DailyCard) ToDaySet(ds *binding.DaySets) error {
	if c.Number < 1 || c.Number > 12 {
		err := fmt.Errorf("неверный номер суточной карты")
		logger.Error.Printf(err.Error())
		return err
	}
	r := binding.NewOneDay(c.Number)
	if len(c.Programs) > 12 {
		err := fmt.Errorf("слишком много переключений суточной карты")
		logger.Error.Printf(err.Error())
		return err
	}
	if c.Programs[0].Hour != 0 || c.Programs[0].Minute != 0 {
		is24_00 := false
		for _, v := range c.Programs {
			if v.Hour == 24 && v.Minute == 0 {
				is24_00 = true
			}
		}
		if !is24_00 {
			err := fmt.Errorf("неверное завершение суточной карты")
			logger.Error.Printf(err.Error())
			return err
		}
	}
	r.Count = len(c.Programs)
	for i := 0; i < r.Count; i++ {
		r.Lines[i].Hour = c.Programs[i].Hour
		r.Lines[i].Min = c.Programs[i].Minute
		r.Lines[i].PKNom = c.Programs[i].Number
	}
	for i := 0; i < len(ds.DaySets); i++ {
		if ds.DaySets[i].Number == r.Number {
			ds.DaySets[i] = r
			return nil
		}
	}
	err := fmt.Errorf("нет суточной карты")
	logger.Error.Printf(err.Error())
	return err
}
func (w *Week) ToWeekSet(dw *binding.WeekSets) error {
	if w.Number < 1 || w.Number > 12 {
		err := fmt.Errorf("неверный номер недельной карты")
		logger.Error.Printf(err.Error())
		return err
	}
	if len(w.DailyCards) < 1 || len(w.DailyCards) > 12 {
		err := fmt.Errorf("неверная длина недельной карты")
		logger.Error.Println(err.Error())
		return err

	}
	ww := binding.OneWeek{Number: w.Number, Days: w.DailyCards}
	for i := 0; i < len(dw.WeekSets); i++ {
		if ww.Number == dw.WeekSets[i].Number {
			dw.WeekSets[i] = ww
			return nil
		}
	}
	err := fmt.Errorf("неверная недельная карта")
	logger.Error.Println(err.Error())
	return err
}
