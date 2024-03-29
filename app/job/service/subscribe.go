package service

import (
	"github.com/micro/go-micro/v2/broker"
	cApi "mercury/app/comet/api"
	"mercury/app/logic/api"
	"mercury/x/ecode"
)

func (s *Service) subscribePushMessage(e broker.Event) error {
	s.log.Info("subscribe", "topic", e.Topic())

	if e.Message() == nil {
		return ecode.NewError("message can not be nil")
	}

	pm := new(api.PushMessage)
	if err := pm.Unmarshal(e.Message().Body); err != nil {
		return err
	}
	if err := s.pushMessage(pm.Operation, pm.ServerID, pm.SIDs, pm.Data); err != nil {
		return err
	}

	return nil
}

func (s *Service) subscribeBroadcastMessage(e broker.Event) error {
	s.log.Info("subscribe", "topic", e.Topic())

	if e.Message() == nil {
		return ecode.NewError("message can not be nil")
	}

	bm := new(api.BroadcastMessage)
	if err := bm.Unmarshal(e.Message().Body); err != nil {
		return err
	}
	if err := s.broadcastMessage(bm.Servers, bm.Data); err != nil {
		return err
	}

	return nil
}

func (s *Service) pushMessage(op int32, serverID string, sids []string, data []byte) error {
	if comet, ok := s.cometServers[serverID]; ok {
		comet.Push(&cApi.PushMessageReq{
			Operation: op,
			SIDs:      sids,
			Data:      data,
		})
	}
	return nil
}

func (s *Service) broadcastMessage(servers map[string]*api.StringSliceValue, data []byte) error {
	if len(servers) > 0 {
		for serverID, value := range servers {
			if comet, ok := s.cometServers[serverID]; ok {
				if value != nil {
					go comet.Push(&cApi.PushMessageReq{
						SIDs: value.Value,
						Data: data,
					})
				}
			}
		}
	} else {
		for _, comet := range s.cometServers {
			go comet.Broadcast(&cApi.BroadcastMessageReq{
				Data: data,
			})
		}
	}
	return nil
}
