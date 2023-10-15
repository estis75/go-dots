package db_models

import "time"
import "gitea.com/xorm/xorm"

type UriFilteringTrafficPerPort struct {
	Id                  int64     `xorm:"'id' pk autoincr"`
	TelePreMitigationId int64     `xorm:"'tele_pre_mitigation_id' not null"`
	TrafficType         string    `xorm:"'traffic_type' enum('TOTAL_TRAFFIC_NORMAL','TOTAL_ATTACK_TRAFFIC','TOTAL_TRAFFIC') not null"`
	Unit                string    `xorm:"'unit' enum('packet-ps','bit-ps','byte-ps','kilopacket-ps','kilobit-ps','kilobytes-ps','megapacket-ps','megabit-ps','megabyte-ps','gigapacket-ps','gigabit-ps','gigabyte-ps','terapacket-ps','terabit-ps','terabyte-ps') not null"`
	Port                int       `xorm:"'port' not null"`
	LowPercentileG      uint64    `xorm:"'low_percentile_g'"`
	MidPercentileG      uint64    `xorm:"'mid_percentile_g'"`
	HighPercentileG     uint64    `xorm:"'high_percentile_g'"`
	PeakG               uint64    `xorm:"'peak_g'"`
	CurrentG            uint64    `xorm:"current_g"`
	Created             time.Time `xorm:"created"`
	Updated             time.Time `xorm:"updated"`
}

// Get uri filtering traffic per port
func GetUriFilteringTrafficPerPort(engine *xorm.Engine, telePreMitigationId int64, trafficType string) (trafficList []UriFilteringTrafficPerPort, err error) {
	trafficList = []UriFilteringTrafficPerPort{}
	err = engine.Where("tele_pre_mitigation_id = ? AND traffic_type = ?", telePreMitigationId, trafficType).OrderBy("id ASC").Find(&trafficList)
	return
}

// Delete uri filtering traffic per port
func DeleteUriFilteringTrafficPerPort(session *xorm.Session, telePreMitigationId int64) (err error) {
	_, err = session.Delete(&UriFilteringTrafficPerPort{TelePreMitigationId: telePreMitigationId})
	return
}