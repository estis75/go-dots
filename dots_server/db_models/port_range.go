package db_models

import (
	"time"

	"github.com/go-xorm/xorm"
)

type PortRange struct {
	Id                int64     `xorm:"'id' pk autoincr"`
	MitigationScopeId int64     `xorm:"'mitigation_scope_id'"`
	LowerPort         int       `xorm:"'lower_port'"`
	UpperPort         int       `xorm:"'upper_port'"`
	Created           time.Time `xorm:"created"`
	Updated           time.Time `xorm:"updated"`
}

func CreatePortRangeParam(lowerPort int, upperPort int) (portRange *PortRange) {
	portRange = new(PortRange)
	portRange.LowerPort = lowerPort
	portRange.UpperPort = upperPort
	return
}

func DeleteMitigationScopePortRange(session *xorm.Session, mitigationScopeId int64) (err error) {
	_, err = session.Delete(&PortRange{MitigationScopeId: mitigationScopeId})
	return
}

