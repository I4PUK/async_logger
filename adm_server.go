package main

import (
	"context"
	"time"
)

type AdminServ struct {
	logger SimpleEventLogger
}

func (adm *AdminServ) mustEmbedUnimplementedAdminServer() {}

func (adm *AdminServ) Logging(n *Nothing, logServerStream Admin_LoggingServer) error {
	ch := adm.logger.Subscribe()

	ctxTimeout, _ := context.WithTimeout(context.Background(), time.Second*3)

	for {
		select {
		case <-ctxTimeout.Done():
			adm.logger.Unsubscribe(ch)
			return nil
		case msg := <-ch:
			err := logServerStream.Send(msg)
			if err != nil {
				adm.logger.Unsubscribe(ch)
				return err
			}
		}
	}

}

func (adm *AdminServ) Statistics(interval *StatInterval, stream Admin_StatisticsServer) error {
	ticker := time.NewTicker(time.Duration(interval.IntervalSeconds) * time.Second)
	ch := adm.logger.Subscribe()
	stat := emptyStat()

	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			adm.logger.Unsubscribe(ch)
			return stream.Context().Err()
		case e := <-ch:
			updateStat(stat, e)
		case <-ticker.C:
			err := stream.Send(stat)
			if err != nil {
				adm.logger.Unsubscribe(ch)
				return err
			}
			stat = emptyStat()
		}
	}
}

func getAdminInstance(logger *SimpleEventLogger) *AdminServ {
	return &AdminServ{
		logger,
	}
}

func emptyStat() *Stat {
	return &Stat{
		ByMethod:   make(map[string]uint64),
		ByConsumer: make(map[string]uint64),
	}
}

func updateStat(stat *Stat, e *Event) {
	stat.Timestamp = time.Now().Unix()
	stat.ByConsumer[e.Consumer]++
	stat.ByMethod[e.Method]++
}
