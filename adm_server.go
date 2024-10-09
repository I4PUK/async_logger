package main

import "time"

type AdminServ struct {
	logger SimpleEventLogger
}

func (adm *AdminServ) mustEmbedUnimplementedAdminServer() {}

func (adm *AdminServ) Logging(n *Nothing, logServerStream Admin_LoggingServer) error {
	ch := adm.logger.Subscribe()

	for e := range ch {
		err := logServerStream.Send(e)
		if err != nil {
			adm.logger.Unsubscribe(ch)
			return err
		}
	}
	return nil
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

func getAdminInstance() *AdminServ {
	return &AdminServ{}
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
