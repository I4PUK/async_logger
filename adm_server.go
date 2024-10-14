package main

import (
	"context"
	"time"
)

type AdminServ struct {
	host   string
	logger *SimpleEventLogger
	stats  *SimpleEventStats
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
	stat := adm.stats.InitStat()

	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			adm.logger.Unsubscribe(ch)
			return stream.Context().Err()
		case e := <-ch:
			adm.stats.UpdateStat(e)
		case <-ticker.C:
			err := stream.Send(stat)
			if err != nil {
				adm.logger.Unsubscribe(ch)
				return err
			}
			adm.stats.InitStat()
		}
	}
}

func getAdminInstance(host string, logger *SimpleEventLogger, stats *SimpleEventStats) *AdminServ {
	return &AdminServ{
		host:   host,
		logger: logger,
		stats:  stats,
	}
}
