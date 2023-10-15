package db_models

import "time"
import "gitea.com/xorm/xorm"

type TelemetryTraffic struct {
	Id              int64     `xorm:"'id' pk autoincr"`
	PrefixType      string    `xorm:"'prefix_type' enum('TARGET_PREFIX','SOURCE_PREFIX') not null"`
	PrefixTypeId    int64     `xorm:"'prefix_type_id' not null"`
	TrafficType     string    `xorm:"'traffic_type' enum('TOTAL_ATTACK_TRAFFIC','TOTAL_TRAFFIC') not null"`
	Unit            string    `xorm:"'unit' enum('packet-ps','bit-ps','byte-ps','kilopacket-ps','kilobit-ps','kilobytes-ps','megapacket-ps','megabit-ps','megabyte-ps','gigapacket-ps','gigabit-ps','gigabyte-ps','terapacket-ps','terabit-ps','terabyte-ps') not null"`
	LowPercentileG  uint64    `xorm:"'low_percentile_g'"`
	MidPercentileG  uint64    `xorm:"'mid_percentile_g'"`
	HighPercentileG uint64    `xorm:"'high_percentile_g'"`
	CurrentG        uint64    `xorm:"current_g"`
	PeakG           uint64    `xorm:"'peak_g'"`
	Created         time.Time `xorm:"created"`
	Updated         time.Time `xorm:"updated"`
}

// Get telemetry traffic (by mitigation)
func GetTelemetryTraffic(engine *xorm.Engine, prefixType string, prefixTypeId int64, trafficType string) (trafficList []TelemetryTraffic, err error) {
	trafficList = []TelemetryTraffic{}
	err = engine.Where("prefix_type = ? AND prefix_type_id = ? AND traffic_type = ?", prefixType, prefixTypeId, trafficType).OrderBy("id ASC").Find(&trafficList)
	return
}

// Delete telemetry traffic (by mitigation)
func DeleteTelemetryTraffic(session *xorm.Session, prefixType string, prefixTypeId int64, trafficType string) (err error) {
	_, err = session.Delete(&TelemetryTraffic{PrefixType: prefixType, PrefixTypeId: prefixTypeId, TrafficType: trafficType})
	return
}