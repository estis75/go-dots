package models

import (
	"time"
	"encoding/json"
	"github.com/go-xorm/xorm"
	"github.com/nttdots/go-dots/dots_common/messages"
	"github.com/nttdots/go-dots/dots_server/db_models"
	log "github.com/sirupsen/logrus"
	dots_config "github.com/nttdots/go-dots/dots_server/config"
	types "github.com/nttdots/go-dots/dots_common/types/data"
	db_models_data "github.com/nttdots/go-dots/dots_server/db_models/data"
)

var DefaultTsid = -1

type SetupType string
const (
	TELEMETRY_CONFIGURATION SetupType = "TELEMETRY_CONFIGURATION"
	PIPE                    SetupType = "PIPE"
	BASELINE                SetupType = "BASELINE"
)

type TelemetryType string
const (
	TELEMETRY       TelemetryType = "TELEMETRY"
	TELEMETRY_SETUP TelemetryType = "TELEMETRY_SETUP"
)

type PrefixType string
const (
	TARGET_PREFIX PrefixType = "TARGET_PREFIX"
	SOURCE_PREFIX PrefixType = "SOURCE_PREFIX"
)

type ParameterType string
const (
	TARGET_PROTOCOL ParameterType = "TARGET_PROTOCOL"
	TARGET_FQDN     ParameterType = "FQDN"
	TARGET_URI      ParameterType = "URI"
	ALIAS_NAME      ParameterType = "ALIAS_NAME"
)

type trafficType string
const (
	TOTAL_TRAFFIC_NORMAL trafficType = "TOTAL_TRAFFIC_NORMAL"
	TOTAL_ATTACK_TRAFFIC trafficType = "TOTAL_ATTACK_TRAFFIC"
	TOTAL_TRAFFIC        trafficType = "TOTAL_TRAFFIC"
)

// Create telemetry configuration
func CreateTelemetryConfiguration(customerId int, cuid string, cdid string, tsid int, telemetryConfiguration *TelemetryConfiguration) (err error) {
	// database connection create
	engine, err := ConnectDB()
	if err != nil {
		log.Errorf("Database connect error: %s", err)
		return err
	}
	// transaction start
	session := engine.NewSession()
	defer session.Close()

	err = session.Begin()
	if err != nil {
		return err
	}

	// Get telemetry setup by 'cuid' and 'setup_type' is 'telemetry_configuration'
	currentSetupList, err := GetTelemetrySetupByCuidAndSetupType(customerId, cuid, string(TELEMETRY_CONFIGURATION))
	if err != nil {
		log.Errorf("Get telemetry setup with setup type is telemetry-configuration err: %+v", err)
		return err
	}
	// If exsited telemetry setup with setup_type is 'telemetry_configuration' in DB, DOTS server will delete it from DB
	for _, currentSetup := range currentSetupList {
		// Deleted current telemetry setup
		err = DeleteTelemetrySetupTelemetryConfiguration(engine, session, currentSetup.Id)
		if err != nil {
			session.Rollback()
			return err
		}
	}

	// Create new telemetry setup configuration
	err = CreateTelemetrySetupTelemetryConfiguration(session, customerId, cuid, cdid, tsid, telemetryConfiguration)
	if err != nil {
		session.Rollback()
		return err
	}
	// add Commit() after all actions
	err = session.Commit()
	return nil
}

// Update telemetry configuration
func UpdateTelemetryConfiguration(customerId int, cuid string, cdid string, tsid int, telemetryConfiguration *TelemetryConfiguration) (err error) {
	// database connection create
	engine, err := ConnectDB()
	if err != nil {
		log.Errorf("Database connect error: %s", err)
		return err
	}
	// transaction start
	session := engine.NewSession()
	defer session.Close()

	err = session.Begin()
	if err != nil {
		return err
	}

	// Get telemetry setup by 'cuid' and 'setup_type' is 'telemetry_configuration'
	currentSetup, err := GetTelemetrySetupByCuidAndSetupType(customerId, cuid, string(TELEMETRY_CONFIGURATION))
	if err != nil {
		log.Errorf("Get telemetry setup with setup type is telemetry-configuration err: %+v", err)
		return err
	}
	if currentSetup[0].Id == 0 {
		// no data found
		log.Debugf("telemetry_setup update data exist err: %s", err)
		return
	}
	// Get current telemetry configuration
	updateTelemetryConfiguration, err := db_models.GetTelemetryConfigurationByTeleSetupId(engine, currentSetup[0].Id)
	if err != nil {
		log.Errorf("Get telemetry configuration err: %+v", err)
		return err
	}
	if updateTelemetryConfiguration.Id == 0 {
		// no data found
		log.Debugf("telemetry_configuration update data exist err: %s", err)
		return
	}

	// Updated telemetry configuration
	updateTelemetryConfiguration.MeasurementInterval = messages.ConvertMeasurementIntervalToString(telemetryConfiguration.MeasurementInterval)
	updateTelemetryConfiguration.MeasurementSample = messages.ConvertMeasurementSampleToString(telemetryConfiguration.MeasurementSample)
	updateTelemetryConfiguration.LowPercentile = telemetryConfiguration.LowPercentile
	updateTelemetryConfiguration.MidPercentile = telemetryConfiguration.MidPercentile
	updateTelemetryConfiguration.HighPercentile = telemetryConfiguration.HighPercentile
	updateTelemetryConfiguration.ServerOriginatedTelemetry = telemetryConfiguration.ServerOriginatedTelemetry
	updateTelemetryConfiguration.TelemetryNotifyInterval = telemetryConfiguration.TelemetryNotifyInterval
	_, err = session.Id(updateTelemetryConfiguration.Id).Update(updateTelemetryConfiguration)
	if err != nil {
		log.Errorf("telemetry_configuration update err: %s", err)
		session.Rollback()
		return err
	}
	// update server_originated_telemetry boolean column
	_, err = session.Id(updateTelemetryConfiguration.Id).Cols("server_originated_telemetry").Update(&updateTelemetryConfiguration)
	if err != nil {
		session.Rollback()
		log.Errorf("telemetry_configuration update err: %s", err)
		return
	}

	// update telemetry_notify_interval column
	_, err = session.Id(updateTelemetryConfiguration.Id).Cols("telemetry_notify_interval").Update(&updateTelemetryConfiguration)
	if err != nil {
		session.Rollback()
		log.Errorf("telemetry_configuration update err: %s", err)
		return
	}

	// Deleted current unit configuration
	err = db_models.DeleteUnitConfigurationByTeleConfigId(session, updateTelemetryConfiguration.Id)
	if err != nil {
		log.Errorf("Delete unit configuration err: %+v", err)
		session.Rollback()
		return err
	}
	// Registered unit configuration
	err = RegisterUnitConfiguration(session, updateTelemetryConfiguration.Id, telemetryConfiguration.UnitConfigList)
	if err != nil {
		return err
	}
	// add Commit() after all actions
	err = session.Commit()
	return nil
}

// Create total pipe capacity
func CreateTotalPipeCapacity(customerId int, cuid string, cdid string, tsid int, pipeList []TotalPipeCapacity, isPresent bool) (isConflict bool, err error) {
	isConflict = false
	// database connection create
	engine, err := ConnectDB()
	if err != nil {
		log.Errorf("Database connect error: %s", err)
		return isConflict, err
	}
	// transaction start
	session := engine.NewSession()
	defer session.Close()

	err = session.Begin()
	if err != nil {
		return isConflict, err
	}

	// Get telemetry setup by customerId and setup_type is 'pipe'
	currentSetupList, err := GetTelemetrySetupByCustomerIdAndSetupType(customerId, string(PIPE))
	if err != nil {
		log.Errorf("Get telemetry setup with setup type is 'pipe' err: %+v", err)
		return isConflict, err
	}
	// If such two requests have overlapping "link-id" and "unit", DOTS server will to do as below:
	// - If DOTS clients are same, DOTS server will delete total pipe capacity with the lower 'tsid' and insert  total pipe capacity with higher 'tsid'
	// - If DOTS clients are difference, DOTS server will response conflict with conflict cause is 'Overlapping targets'
	for _, currentSetup := range currentSetupList {
		// Get total pipe capacity by teleSetupId
		currentPipeList, err := db_models.GetTotalPipeCapacityByTeleSetupId(engine, currentSetup.Id)
		if err != nil {
			log.Errorf("Get total-pipe-capacity err: %+v", err)
			return isConflict, err
		}
		lenCurrentPipeList := len(currentPipeList)
		for _, pipe := range pipeList {
			for _, currentPipe := range currentPipeList {
				if (pipe.LinkId == currentPipe.LinkId && pipe.Capacity == 0) || currentSetup.Tsid == DefaultTsid {
					// Delete total pipe capacity
					log.Debugf("[Capacity is 0] The request capacity is 0. DOTS server will delete total_pipe_capacity id = %+v", currentPipe.Id)
					err = db_models.DeleteTotalPipeCapacityById(session, currentPipe.Id)
					if err != nil {
						log.Errorf("Delete total pipe capacity err: %+v", err)
						session.Rollback()
						return isConflict, err
					}
					lenCurrentPipeList--
				} else if pipe.LinkId == currentPipe.LinkId && messages.ConvertUnitToString(pipe.Unit) == currentPipe.Unit {
					if currentSetup.Cuid == cuid && currentSetup.Tsid < tsid {
						log.Debugf("[Overlap] Overlapping link_id = %+v and unit_id = %+v. DOTS server will delete total_pipe_capacity id = %+v", 
						currentPipe.LinkId, currentPipe.Unit, currentPipe.Id)
						// Delete total pipe capacity
						err = db_models.DeleteTotalPipeCapacityById(session, currentPipe.Id)
						if err != nil {
							log.Errorf("Delete total pipe capacity err: %+v", err)
							session.Rollback()
							return isConflict, err
						}
						lenCurrentPipeList--
					} else if currentSetup.Cuid != cuid {
						// Set isConflict to true
						isConflict = true
						return isConflict, nil
					}
				}
			}
		}
		if lenCurrentPipeList == 0 {
			// Deleted current telemetry setup
			err = db_models.DeleteTelemetrySetupById(session, currentSetup.Id)
			if err != nil {
				log.Errorf("Delete telemetry setup error: %s", err)
				session.Rollback()
				return isConflict, err
			}
		}
	}
	if !isPresent {
		// Create telemetry setup with setup_type is 'pipe'
		log.Debug("Create total pipe capacity")
		err = CreateTelemetrySetupPipe(session, customerId, cuid, cdid, tsid, pipeList)
		if err != nil {
			session.Rollback()
			return isConflict, err
		}
	} else {
		log.Debug("Update total pipe capacity")
		// Update telemetry setup with setup_type is 'pipe'
		err = UpdateTotalPipeCapacity(session, customerId, cuid, cdid, tsid, pipeList)
		if err != nil {
			session.Rollback()
			return isConflict, err
		}
	}
	// add Commit() after all actions
	err = session.Commit()
	return isConflict, err
}

// Update total pipe capacity
func UpdateTotalPipeCapacity(session *xorm.Session, customerId int, cuid string, cdid string, tsid int, pipeList []TotalPipeCapacity) error {
	// Get telemetry setup by tsid and setup_type is 'pipe'
	currentSetup, err := GetTelemetrySetupByTsidAndSetupType(customerId, cuid, tsid, string(PIPE))
	if err != nil {
		log.Errorf("Get telemetry setup with setup type is 'pipe' err: %+v", err)
		return err
	}
	if currentSetup.Id == 0 {
		// no data found 
		log.Debugf("telemetry setup update data exist err: %s", err)
		return nil
	}
	// Get current total pipe capacity by teleSetupId
	currentPipeList, err := db_models.GetTotalPipeCapacityByTeleSetupId(engine, currentSetup.Id)
	if err != nil {
		log.Errorf("Get total pipe capacity err: %+v", err)
		return err
	}
	for _, currentPipe := range currentPipeList {
		// Delete total pipe capacity
		err = db_models.DeleteTotalPipeCapacityById(session, currentPipe.Id)
		if err != nil {
			log.Errorf("Delete total pipe capacity err: %+v", err)
			return err
		}
	}
	// Registered total pipe capacity
	err = RegisterTotalPipecapacity(session, currentSetup.Id, pipeList)
	if err != nil {
		return err
	}
	return nil
}

// Create baseline
func CreateBaseline(customerId int, cuid string, cdid string, tsid int, baselineList []Baseline, isPresent bool) (isConflict bool, err error) {
	isConflict = false
	// database connection create
	engine, err := ConnectDB()
	if err != nil {
		log.Errorf("Database connect error: %s", err)
		return isConflict, err
	}
	// transaction start
	session := engine.NewSession()
	defer session.Close()

	err = session.Begin()
	if err != nil {
		return isConflict, err
	}

	// Get telemetry setup by customerId and setup_type is 'baseline'
	currentSetupList, err := GetTelemetrySetupByCustomerIdAndSetupType(customerId, string(BASELINE))
	if err != nil {
		log.Errorf("Get telemetry setup with setup type is 'baseline' err: %+v", err)
		return isConflict, err
	}
	// If such two requests have overlapping targets ('target-prefix', 'target-uri', 'taget-fqdn'), DOTS server will to do as below:
	// - If DOTS clients are same, DOTS server will delete baseline with the lower 'tsid' and insert  baseline with higher 'tsid'
	// - If DOTS clients are difference, DOTS server will response conflict with conflict cause is 'Overlapping targets'
	for _, currentSetup := range currentSetupList {
		// Get baseline
		currentBaselineList, err := GetBaselineByTeleSetupId(customerId, currentSetup.Cuid, currentSetup.Id)
		if err != nil {
			log.Errorf("Get baseline by teleSetupId err: %+v", err)
			return isConflict, err
		}
		lenCurrentBaselineList := len(currentBaselineList)
		for _, baseline := range baselineList {
			for _, currentBaseline := range currentBaselineList {
				if currentSetup.Cuid == cuid && currentSetup.Tsid == tsid {
					continue
				}
				// Check overlapping targets
				isOverlap := CheckOverlapTargetList(baseline.TargetList, currentBaseline.TargetList)
				if (isOverlap && currentSetup.Cuid == cuid && currentSetup.Tsid < tsid) || currentSetup.Tsid == DefaultTsid {
					// Delete baseline
					log.Debugf("DOTS server will delete baseline id = %+v", currentBaseline.Id)
					err = DeleteBaseline(session, currentBaseline.Id)
					if err != nil {
						session.Rollback()
						return isConflict, err
					}
					lenCurrentBaselineList--
				} else if isOverlap && currentSetup.Cuid != cuid  {
					isConflict = true
					return isConflict, nil
				}
			}
		}
		if lenCurrentBaselineList == 0 {
			// Deleted current telemetry setup
			err = db_models.DeleteTelemetrySetupById(session, currentSetup.Id)
			if err != nil {
				log.Errorf("Delete telemetry setup error: %s", err)
				session.Rollback()
				return isConflict, err
			}
		}
	}
	if !isPresent {
		// Create telemetry setup with setup_type is 'baseline'
		log.Debug("Create baseline")
		err = CreateTelemetrySetupBaseline(session, customerId, cuid, cdid, tsid, baselineList)
		if err != nil {
			session.Rollback()
			return isConflict, err
		}
	} else {
		// Update telemetry setup with setup_type is 'baseline'
		log.Debug("Update baseline")
		err = UpdateBaseline(session, customerId, cuid, cdid, tsid, baselineList)
		if err != nil {
			session.Rollback()
			return isConflict, err
		}
	}
	// add Commit() after all actions
	err = session.Commit()
	return isConflict, err
}

// Update baseline
func UpdateBaseline(session *xorm.Session, customerId int, cuid string, cdid string, tsid int, baselineList []Baseline) error {
	// Get telemetry setup by tsid and setup_type is 'baseline'
	currentSetup, err := GetTelemetrySetupByTsidAndSetupType(customerId, cuid, tsid, string(BASELINE))
	if err != nil {
		log.Errorf("Get telemetry setup with setup type is 'baseline' err: %+v", err)
		return err
	}
	if currentSetup.Id == 0 {
		// no data found 
		log.Debugf("telemetry setup update data exist err: %s", err)
		return nil
	}
	// Get current baseline by teleSetupId
	currentBaselineList, err := db_models.GetBaselineByTeleSetupId(engine, currentSetup.Id)
	if err != nil {
		log.Errorf("Get baseline by teleSetupId err: %+v", err)
		return err
	}
	// If existed baseline in DB, DOTS server will delete current baseline and create new baseline
	for _, currentBaseline := range currentBaselineList {
		// Delete baseline
		err = DeleteBaseline(session, currentBaseline.Id)
		if err != nil {
			return err
		}
	}
	// Create baseline
	err = createBaseline(session, currentSetup.Id, baselineList)
	if err != nil {
		return err
	}
	return nil
}

// Create telemetry setup with setup_type is 'telemetry_configuration'
func CreateTelemetrySetupTelemetryConfiguration(session *xorm.Session, customerId int, cuid string, cdid string, tsid int, telemetryConfiguration *TelemetryConfiguration) error {
	// Registered telemetry setup with setup type is telemetry configuration
	newTelemetrySetup, err := RegisterTelemetrySetup(session, customerId, cuid, cdid, tsid, string(TELEMETRY_CONFIGURATION))
	if err != nil {
		return err
	}
	// Create telemetry configuration
	err = createTelemetryConfiguration(session, newTelemetrySetup.Id, telemetryConfiguration)
	if err != nil {
		return err
	}
	return nil
}

// Create telemetry setup with setup_type is 'pipe'
func CreateTelemetrySetupPipe(session *xorm.Session, customerId int, cuid string, cdid string, tsid int, pipeList []TotalPipeCapacity) error {
	// Registered telemetry setup with setup type is pipe
	newTelemetrySetup, err := RegisterTelemetrySetup(session, customerId, cuid, cdid, tsid, string(PIPE))
	if err != nil {
		return err
	}
	// Registered total pipe capacity
	err = RegisterTotalPipecapacity(session, newTelemetrySetup.Id, pipeList)
	if err != nil {
		return err
	}
	return nil
}

// Create telemetry setup with setup_type is 'baseline'
func CreateTelemetrySetupBaseline(session *xorm.Session, customerId int, cuid string, cdid string, tsid int, baselineList []Baseline) error {
	// Registered telemetry setup with setup type is baseline
	newTelemetrySetup, err := RegisterTelemetrySetup(session, customerId, cuid, cdid, tsid, string(BASELINE))
	if err != nil {
		return err
	}
	// Create baseline (targets, total_traffic_normal, total_traffic_normal_per_protocol, total_traffic_normal_per_port, total_connection_capacity, total_connection_capacity_per_port)
	err = createBaseline(session, newTelemetrySetup.Id, baselineList)
	if err != nil {
		return err
	}
	return nil
}

// Create telemetry configuration
func createTelemetryConfiguration(session *xorm.Session, teleSetupId int64, telemetryConfiguration *TelemetryConfiguration) error {
	// Registered telemetry configuration
	newTelemetryConfiguration, err := RegisterTelemetryConfiguration(session, teleSetupId, telemetryConfiguration)
	if err != nil {
		return err
	}
	// Registered unit configuration
	err = RegisterUnitConfiguration(session, newTelemetryConfiguration.Id, telemetryConfiguration.UnitConfigList)
	if err != nil {
		return err
	}
	return nil
}

// Create baseline
func createBaseline(session *xorm.Session, teleSetupId int64, baselines []Baseline)  error {
	for _, baseline := range baselines {
		// Registered baseline
		newBaseline, err := RegisterBaseline(session, teleSetupId, baseline)
		if err != nil {
			return err
		}
		// Registered telemetry prefix
		err = RegisterTelemetryPrefix(session, string(TELEMETRY_SETUP), newBaseline.Id, string(TARGET_PREFIX), baseline.TargetPrefix)
		if err != nil {
			return err
		}
		// Registered telemetry port range
		err = RegisterTelemetryPortRange(session, string(TELEMETRY_SETUP), newBaseline.Id, string(TARGET_PREFIX), baseline.TargetPortRange)
		if err != nil {
			return err
		}
		// Create telemetry parameter value
		err = CreateTelemetryParameterValue(session, string(TELEMETRY_SETUP), newBaseline.Id, baseline.TargetProtocol, baseline.FQDN, baseline.URI, baseline.AliasName)
		if err != nil {
			return err
		}
		// Registered total traffic normal
		err = RegisterTraffic(session, string(TELEMETRY_SETUP), string(TARGET_PREFIX), newBaseline.Id, string(TOTAL_TRAFFIC_NORMAL), baseline.TotalTrafficNormal)
		if err != nil {
			return err
		}
		// Registered total traffic normal per protocol
		err = RegisterTrafficPerProtocol(session, string(TELEMETRY_SETUP), newBaseline.Id, string(TOTAL_TRAFFIC_NORMAL), baseline.TotalTrafficNormalPerProtocol)
		if err != nil {
			return err
		}
		// Registered total traffic normal per port
		err = RegisterTrafficPerPort(session, string(TELEMETRY_SETUP), newBaseline.Id, string(TOTAL_TRAFFIC_NORMAL), baseline.TotalTrafficNormalPerPort)
		if err != nil {
			return err
		}
		// Registered total connection capacity
		err = RegisterTotalConnectionCapacity(session, newBaseline.Id, baseline.TotalConnectionCapacity)
		if err != nil {
			return err
		}
		// Registered total connection capacity per port
		err = RegisterTotalConnectionCapacityPerPort(session, newBaseline.Id, baseline.TotalConnectionCapacityPerPort)
		if err != nil {
			return err
		}
	}
	return nil
}

// Registed telemetry setup to DB
func RegisterTelemetrySetup(session *xorm.Session, customerId int, cuid string, cdid string, tsid int, setupType string) (*db_models.TelemetrySetup, error) {
	newTelemetrySetup := db_models.TelemetrySetup{
		CustomerId: customerId,
		Cuid:       cuid,
		Cdid:       cdid,
		Tsid:       tsid,
		SetupType:  setupType,
	}
	_, err := session.Insert(&newTelemetrySetup)
	if err != nil {
		log.Errorf("telemetry setup insert err: %s", err)
		return nil, err
	}
	return &newTelemetrySetup , nil
}

// Registered telemetry configuration to DB
func RegisterTelemetryConfiguration(session *xorm.Session, teleSetupId int64, telemetryConfiguration *TelemetryConfiguration) (*db_models.TelemetryConfiguration, error) {
	newTelemetryConfiguration := db_models.TelemetryConfiguration{
		TeleSetupId:               teleSetupId,
		MeasurementInterval:       messages.ConvertMeasurementIntervalToString(telemetryConfiguration.MeasurementInterval),
		MeasurementSample:         messages.ConvertMeasurementSampleToString(telemetryConfiguration.MeasurementSample),
		LowPercentile:             telemetryConfiguration.LowPercentile,
		MidPercentile:             telemetryConfiguration.MidPercentile,
		HighPercentile:            telemetryConfiguration.HighPercentile,
		ServerOriginatedTelemetry: telemetryConfiguration.ServerOriginatedTelemetry,
		TelemetryNotifyInterval:   telemetryConfiguration.TelemetryNotifyInterval,
	}
	_, err := session.Insert(&newTelemetryConfiguration)
	if err != nil {
		log.Errorf("telemetry configuration insert err: %s", err)
		return nil, err
	}
	return &newTelemetryConfiguration, nil
}

// Registed total pipe capacity to DB
func RegisterTotalPipecapacity(session *xorm.Session, teleSetupId int64, pipeList []TotalPipeCapacity) error {
	newTotalPipeCapacityList := []db_models.TotalPipeCapacity{}
	for _, pipe := range pipeList {
		if pipe.Capacity > 0 {
			newTotalPipeCapacity := db_models.TotalPipeCapacity{
				TeleSetupId: teleSetupId,
				LinkId:      pipe.LinkId,
				Capacity:    uint64(pipe.Capacity),
				Unit:        messages.ConvertUnitToString(pipe.Unit),
			}
			newTotalPipeCapacityList = append(newTotalPipeCapacityList, newTotalPipeCapacity)
		}
	}
	if len(newTotalPipeCapacityList) > 0 {
		_, err := session.Insert(&newTotalPipeCapacityList)
		if err != nil {
			log.Errorf("total pipe capacity insert err: %s", err)
			return err
		}
	} else {
		// Deleted current telemetry setup
		err := db_models.DeleteTelemetrySetupById(session, teleSetupId)
		if err != nil {
			log.Errorf("Delete telemetry setup error: %s", err)
			return err
		}
	}
	return nil
}

// Registed baseline to DB
func RegisterBaseline(session *xorm.Session, teleSetupId int64, baseline Baseline) (*db_models.Baseline, error) {
	newBaseline := db_models.Baseline{
		TeleSetupId: teleSetupId,
		BaselineId:  baseline.BaselineId,
	}
	_, err := session.Insert(&newBaseline)
	if err != nil {
		log.Errorf("baseline insert err: %s", err)
		return nil, err
	}
	return &newBaseline, nil
}

// Registered unit configuration to DB
func RegisterUnitConfiguration(session *xorm.Session, tConID int64, unitConfigList []UnitConfig) (err error) {
	newUnitConfigList := []db_models.UnitConfiguration{}
	for _, config := range unitConfigList {
		unit := messages.ConvertUnitToString(config.Unit)
		newUnitConfig := db_models.CreateUnitConfiguration(tConID, unit, config.UnitStatus)
		newUnitConfigList = append(newUnitConfigList, *newUnitConfig)
	}

	if len(newUnitConfigList) > 0 {
		_, err = session.Insert(&newUnitConfigList)
		if err != nil {
			log.Errorf("unit configuration insert err: %s", err)
			return err
		}
	}
	return
}

// Registered telemetry prefix to DB
func RegisterTelemetryPrefix(session *xorm.Session, tType string, typeId int64, prefixType string, prefixs []Prefix) error {
	newTelemetryPrefixList := []db_models.TelemetryPrefix{}
	for _, prefix := range prefixs {
		newTelemetryPrefix := db_models.TelemetryPrefix{
			Type:       tType,
			TypeId:     typeId,
			PrefixType: prefixType,
			Addr:       prefix.Addr,
			PrefixLen:  prefix.PrefixLen,
		}
		newTelemetryPrefixList = append(newTelemetryPrefixList, newTelemetryPrefix)
	}
	if len(newTelemetryPrefixList) > 0 {
		_, err := session.Insert(&newTelemetryPrefixList)
		if err != nil {
			log.Errorf("telemetry prefix insert err: %s", err)
			return err
		}
	}
	return nil
}

// Registed telemetry port range to DB
func RegisterTelemetryPortRange(session *xorm.Session, tType string, typeId int64, prefixType string, portRanges []PortRange) error {
	newTelemetryPortRangeList := []db_models.TelemetryPortRange{}
	for _, portRange := range portRanges {
		newTelemetryPortRange := db_models.TelemetryPortRange{
			Type:       tType,
			TypeId:     typeId,
			PrefixType: prefixType,
			LowerPort:  portRange.LowerPort,
			UpperPort:  portRange.UpperPort,
		}
		newTelemetryPortRangeList = append(newTelemetryPortRangeList, newTelemetryPortRange)
	}
	if len(newTelemetryPortRangeList) > 0 {
		_, err := session.Insert(&newTelemetryPortRangeList)
		if err != nil {
			log.Errorf("telemetry port range insert err: %s", err)
			return err
		}
	}
	return nil
}

// Registed telemetry parameter value (target-protocol, target-fqdn, target-uri) to DB
func CreateTelemetryParameterValue(session *xorm.Session, tType string, typeId int64, protocols SetInt, fqdns SetString, uris SetString, aliasNames SetString) error {
	// Registered protocol to DB
	err := RegisterTelemetryParameterIntValue(session, tType, typeId, string(TARGET_PROTOCOL), protocols)
	if err != nil {
		return err
	}
	// Registered fqdn to DB
	err = RegisterTelemetryParameterStringValue(session, tType, typeId, string(TARGET_FQDN), fqdns)
	if err != nil {
		return err
	}
	// Registered uri to DB
	err = RegisterTelemetryParameterStringValue(session, tType, typeId, string(TARGET_URI), uris)
	if err != nil {
		return err
	}
	// Registered alias-name to DB
	err = RegisterTelemetryParameterStringValue(session, tType, typeId, string(ALIAS_NAME), aliasNames)
	if err != nil {
		return err
	}
	return nil
}

// Registered telemetry parameter string value
func RegisterTelemetryParameterStringValue(session *xorm.Session, tType string, typeId int64, parameterType string, stringValues SetString) error {
	newTeleParameterList := []db_models.TelemetryParameterValue{}
	for _, stringValue := range stringValues.List() {
		if stringValue == "" {
			continue
		}
		newTeleParameter := db_models.TelemetryParameterValue{
			Type:          tType,
			TypeId:        typeId,
			ParameterType: parameterType,
			StringValue:   stringValue,
		}
		newTeleParameterList = append(newTeleParameterList, newTeleParameter)
	}
	if len(newTeleParameterList) > 0 {
		_, err := session.Insert(&newTeleParameterList)
		if err != nil {
			log.Errorf("telemetry parameter value insert err: %s", err)
			return err
		}
	}
	return nil
}
// Registered telemetry parameter int value
func RegisterTelemetryParameterIntValue(session *xorm.Session, tType string, typeId int64, parameterType string, intValues SetInt) error {
	newTeleParameterList := []db_models.TelemetryParameterValue{}
	for _, intValue := range intValues.List() {
		newTeleParameter := db_models.TelemetryParameterValue{
			Type:          tType,
			TypeId:        typeId,
			ParameterType: parameterType,
			IntValue:      intValue,
		}
		newTeleParameterList = append(newTeleParameterList, newTeleParameter)
	}
	if len(newTeleParameterList) > 0 {
		_, err := session.Insert(&newTeleParameterList)
		if err != nil {
			log.Errorf("telemetry parameter value insert err: %s", err)
			return err
		}
	}
	return nil
}

// Registered traffic to DB
func RegisterTraffic(session *xorm.Session, tType string, prefixType string, typeId int64, trafficType string, traffics []Traffic) error {
	newTrafficList := []db_models.Traffic{}
	for _, vTraffic := range traffics {
		newTraffic := db_models.Traffic{
			Type:            tType,
			PrefixType:      prefixType,
			TypeId:          typeId,
			TrafficType:     trafficType,
			Unit:            messages.ConvertUnitToString(vTraffic.Unit),
			LowPercentileG:  uint64(vTraffic.LowPercentileG),
			MidPercentileG:  uint64(vTraffic.MidPercentileG),
			HighPercentileG: uint64(vTraffic.HighPercentileG),
			PeakG:           uint64(vTraffic.PeakG),
		}
		newTrafficList = append(newTrafficList, newTraffic)
	}
	if len(newTrafficList) > 0 {
		_, err := session.Insert(&newTrafficList)
		if err != nil {
			log.Errorf("traffic insert err: %s", err)
			return err
		}
	}
	return nil
}

// Registered traffic per procol to DB
func RegisterTrafficPerProtocol(session *xorm.Session, tType string, typeId int64, trafficType string, traffics []TrafficPerProtocol) error {
	newTrafficList := []db_models.TrafficPerProtocol{}
	for _, vTraffic := range traffics {
		newTraffic := db_models.TrafficPerProtocol{
			Type:            tType,
			TypeId:          typeId,
			TrafficType:     trafficType,
			Unit:            messages.ConvertUnitToString(vTraffic.Unit),
			Protocol:        vTraffic.Protocol,
			LowPercentileG:  uint64(vTraffic.LowPercentileG),
			MidPercentileG:  uint64(vTraffic.MidPercentileG),
			HighPercentileG: uint64(vTraffic.HighPercentileG),
			PeakG:           uint64(vTraffic.PeakG),
		}
		newTrafficList = append(newTrafficList, newTraffic)
	}
	if len(newTrafficList) > 0 {
		_, err := session.Insert(&newTrafficList)
		if err != nil {
			log.Errorf("traffic per protocol insert err: %s", err)
			return err
		}
	}
	return nil
}

// Registered traffic per port to DB
func RegisterTrafficPerPort(session *xorm.Session, tType string, typeId int64, trafficType string, traffics []TrafficPerPort) error {
	newTrafficList := []db_models.TrafficPerPort{}
	for _, vTraffic := range traffics {
		newTraffic := db_models.TrafficPerPort{
			Type:            tType,
			TypeId:          typeId,
			TrafficType:     trafficType,
			Unit:            messages.ConvertUnitToString(vTraffic.Unit),
			Port:            vTraffic.Port,
			LowPercentileG:  uint64(vTraffic.LowPercentileG),
			MidPercentileG:  uint64(vTraffic.MidPercentileG),
			HighPercentileG: uint64(vTraffic.HighPercentileG),
			PeakG:           uint64(vTraffic.PeakG),
		}
		newTrafficList = append(newTrafficList, newTraffic)
	}
	if len(newTrafficList) > 0 {
		_, err := session.Insert(&newTrafficList)
		if err != nil {
			log.Errorf("traffic per port insert err: %s", err)
			return err
		}
	}
	return nil
}

// Registered total connection capacity to DB
func RegisterTotalConnectionCapacity(session *xorm.Session, teleBaselineId int64, tccs []TotalConnectionCapacity) error {
	newTccList := []db_models.TotalConnectionCapacity{}
	for _, vTcc := range tccs {
		newTcc  := db_models.TotalConnectionCapacity {
			TeleBaselineId:          teleBaselineId,
			Protocol:                vTcc.Protocol,
			Connection:              uint64(vTcc.Connection),
			ConnectionClient:        uint64(vTcc.ConnectionClient),
			Embryonic:               uint64(vTcc.Embryonic),
			EmbryonicClient:         uint64(vTcc.EmbryonicClient),
			ConnectionPs:            uint64(vTcc.ConnectionPs),
			ConnectionClientPs:      uint64(vTcc.ConnectionClientPs),
			RequestPs:               uint64(vTcc.RequestPs),
			RequestClientPs:         uint64(vTcc.RequestClientPs),
			PartialRequestMax:       uint64(vTcc.PartialRequestMax),
			PartialRequestClientMax: uint64(vTcc.PartialRequestClientMax),
		}
		newTccList = append(newTccList, newTcc)
	}
	if len(newTccList) > 0 {
		_, err := session.Insert(&newTccList)
		if err != nil {
			log.Errorf("total connection capacity insert err: %s", err)
			return err
		}
	}
	return nil
}

// Registered total connection capacity per port to DB
func RegisterTotalConnectionCapacityPerPort(session *xorm.Session, teleBaselineId int64, tccs []TotalConnectionCapacityPerPort) error {
	newTccList := []db_models.TotalConnectionCapacityPerPort{}
	for _, vTcc := range tccs {
		newTcc  := db_models.TotalConnectionCapacityPerPort {
			TeleBaselineId:          teleBaselineId,
			Protocol:                vTcc.Protocol,
			Port:                    vTcc.Port,
			Connection:              uint64(vTcc.Connection),
			ConnectionClient:        uint64(vTcc.ConnectionClient),
			Embryonic:               uint64(vTcc.Embryonic),
			EmbryonicClient:         uint64(vTcc.EmbryonicClient),
			ConnectionPs:            uint64(vTcc.ConnectionPs),
			ConnectionClientPs:      uint64(vTcc.ConnectionClientPs),
			RequestPs:               uint64(vTcc.RequestPs),
			RequestClientPs:         uint64(vTcc.RequestClientPs),
			PartialRequestMax:       uint64(vTcc.PartialRequestMax),
			PartialRequestClientMax: uint64(vTcc.PartialRequestClientMax),
		}
		newTccList = append(newTccList, newTcc)
	}
	if len(newTccList) > 0 {
		_, err := session.Insert(&newTccList)
		if err != nil {
			log.Errorf("total connection capacity per port insert err: %s", err)
			return err
		}
	}
	return nil
}

// Get telemetry configuration
func GetTelemetryConfiguration(teleSetupId int64) (telemetryConfiguration *TelemetryConfiguration, err error) {
	// database connection create
	engine, err := ConnectDB()
	if err != nil {
		log.Errorf("Database connect error: %s", err)
		return nil, err
	}

	// Get telemetry configuration table data
	dbTelemetryConfiguration, err := db_models.GetTelemetryConfigurationByTeleSetupId(engine, teleSetupId)
	if err != nil {
		log.Error("Get telemetry_configuration by teleSetupId err: %+v", err)
		return nil, err
	}
	telemetryConfiguration = &TelemetryConfiguration{}
	telemetryConfiguration.MeasurementInterval       = messages.ConvertMeasurementIntervalToInt(dbTelemetryConfiguration.MeasurementInterval)
	telemetryConfiguration.MeasurementSample         = messages.ConvertMeasurementSampleToInt(dbTelemetryConfiguration.MeasurementSample)
	telemetryConfiguration.LowPercentile             = dbTelemetryConfiguration.LowPercentile
	telemetryConfiguration.MidPercentile             = dbTelemetryConfiguration.MidPercentile
	telemetryConfiguration.HighPercentile            = dbTelemetryConfiguration.HighPercentile
	telemetryConfiguration.ServerOriginatedTelemetry = dbTelemetryConfiguration.ServerOriginatedTelemetry
	telemetryConfiguration.TelemetryNotifyInterval   = dbTelemetryConfiguration.TelemetryNotifyInterval

	// Get unit configuration data
	dbUnitConfigurationList := []db_models.UnitConfiguration{}
	err = engine.Where("tele_config_id = ?", dbTelemetryConfiguration.Id).OrderBy("id ASC").Find(&dbUnitConfigurationList)
	if err != nil {
		return
	}
	for _, v := range dbUnitConfigurationList {
		unitConfig := UnitConfig{}
		unitConfig.Unit       = messages.ConvertUnitToInt(v.Unit)
		unitConfig.UnitStatus = v.UnitStatus
		telemetryConfiguration.UnitConfigList = append(telemetryConfiguration.UnitConfigList, unitConfig)
	}
	return
}

// Get total pipe capacity
func GetTotalPipeCapacity(teleSetupId int64) (totalPipeCapacityList []TotalPipeCapacity, err error) {
	// database connection create
	engine, err := ConnectDB()
	if err != nil {
		log.Errorf("Database connect error: %s", err)
		return nil, err
	}

	// Get total pipe capacity table data
	dbPipeList, err := db_models.GetTotalPipeCapacityByTeleSetupId(engine, teleSetupId)
	if err != nil {
		log.Error("Get total_pipe_capacity by teleSetupId err: %+v", err)
		return nil, err
	}
	totalPipeCapacityList = []TotalPipeCapacity{}
	for _, dbPipe := range dbPipeList {
		totalPipeCapacity := TotalPipeCapacity{}
		totalPipeCapacity.LinkId   = dbPipe.LinkId
		totalPipeCapacity.Capacity = messages.Uint64String(dbPipe.Capacity)
		totalPipeCapacity.Unit     = messages.ConvertUnitToInt(dbPipe.Unit)
		totalPipeCapacityList      = append(totalPipeCapacityList, totalPipeCapacity)
	}
	return totalPipeCapacityList, nil
}

// Get telemetry setup by cuid and setup type (telemetry_configuration, pipe, baseline)
func GetTelemetrySetupByCuidAndSetupType(customerId int, cuid string, setupType string) (dbTelemetrySetupList []db_models.TelemetrySetup, err error) {
	// database connection create
	engine, err := ConnectDB()
	if err != nil {
		log.Errorf("Database connect error: %s", err)
		return
	}
	dbTelemetrySetupList = []db_models.TelemetrySetup{}
	err = engine.Where("customer_id = ? AND cuid = ? AND setup_type = ?", customerId, cuid, setupType).Find(&dbTelemetrySetupList)
	if err != nil {
		return
	}
	return
}

// Get telemetry setup by cuid
func GetTelemetrySetupByCuid(customerId int, cuid string) (dbTelemetrySetupList []db_models.TelemetrySetup, err error) {
	// database connection create
	engine, err := ConnectDB()
	if err != nil {
		log.Errorf("Database connect error: %s", err)
		return
	}
	dbTelemetrySetupList = []db_models.TelemetrySetup{}
	err = engine.Where("customer_id = ? AND cuid = ?", customerId, cuid).Find(&dbTelemetrySetupList)
	if err != nil {
		return
	}
	return
}

// Get telemetry setup by cuid and tsid >= 0
func GetTelemetrySetupByCuidAndNonNegativeTsid(customerId int, cuid string) (dbTelemetrySetupList []db_models.TelemetrySetup, err error) {
	// database connection create
	engine, err := ConnectDB()
	if err != nil {
		log.Errorf("Database connect error: %s", err)
		return
	}
	dbTelemetrySetupList = []db_models.TelemetrySetup{}
	err = engine.Where("customer_id = ? AND cuid = ? AND tsid >= ?", customerId, cuid, 0).Find(&dbTelemetrySetupList)
	if err != nil {
		return
	}
	return
}

// Get telemetry setup by customerId and setup type (telemetry_configuration, pipe, baseline)
func GetTelemetrySetupByCustomerIdAndSetupType(customerId int, setupType string) (dbTelemetrySetupList []db_models.TelemetrySetup, err error) {
	// database connection create
	engine, err := ConnectDB()
	if err != nil {
		log.Errorf("Database connect error: %s", err)
		return
	}
	dbTelemetrySetupList = []db_models.TelemetrySetup{}
	err = engine.Where("customer_id = ? AND setup_type = ?", customerId, setupType).Find(&dbTelemetrySetupList)
	if err != nil {
		return
	}
	return
}

// Get telemetry setup by tsid and setup type (telemetry_configuration, pipe, baseline)
func GetTelemetrySetupByTsidAndSetupType(customerId int, cuid string, tsid int, setupType string) (dbTelemetrySetup db_models.TelemetrySetup, err error) {
	// database connection create
	engine, err := ConnectDB()
	if err != nil {
		log.Errorf("Database connect error: %s", err)
		return
	}
	dbTelemetrySetup = db_models.TelemetrySetup{}
	_, err = engine.Where("customer_id = ? AND cuid = ? AND tsid = ? AND setup_type = ?", customerId, cuid, tsid, setupType).Get(&dbTelemetrySetup)
	if err != nil {
		return
	}
	return
}

// Get telemetry setup by tsid
func GetTelemetrySetupByTsid(customerId int, cuid string, tsid int) (dbTelemetrySetupList []db_models.TelemetrySetup, err error) {
	// database connection create
	engine, err := ConnectDB()
	if err != nil {
		log.Errorf("Database connect error: %s", err)
		return
	}
	dbTelemetrySetupList = []db_models.TelemetrySetup{}
	err = engine.Where("customer_id = ? AND cuid = ? AND tsid = ?", customerId, cuid, tsid).Find(&dbTelemetrySetupList)
	if err != nil {
		return
	}
	return
}

// Get baseline by teleSetupId
func GetBaselineByTeleSetupId(customerId int, cuid string, setupId int64) (baselineList []Baseline, err error) {
	// database connection create
	engine, err := ConnectDB()
	if err != nil {
		log.Errorf("Database connect error: %s", err)
		return nil, err
	}
	baselineList = []Baseline{}
	// Get baseline by teleSetupId
	baselines, err := db_models.GetBaselineByTeleSetupId(engine, setupId)
	if err != nil {
		log.Errorf("Get baseline err: %+v", err)
		return nil, err
	}
	for _, vBaseline := range baselines {
		baseline := Baseline{}
		baseline.Id = vBaseline.Id
		baseline.BaselineId = vBaseline.BaselineId
		// Get telemetry prefix
		prefixList, err := GetTelemetryPrefix(engine, string(TELEMETRY_SETUP), vBaseline.Id, string(TARGET_PREFIX))
		if err != nil {
			return nil, err
		}
		baseline.TargetPrefix = prefixList
		// Get telemetry port range
		portRangeList, err := GetTelemetryPortRange(engine, string(TELEMETRY_SETUP), vBaseline.Id, string(TARGET_PREFIX))
		if err != nil {
			return nil, err
		}
		baseline.TargetPortRange = portRangeList
		// Get telemetry parameter value with parameter type is 'protocol'
		protocolList, err := GetTelemetryParameterWithParameterTypeIsProtocol(engine, string(TELEMETRY_SETUP), vBaseline.Id, string(TARGET_PROTOCOL))
		if err != nil {
			return nil, err
		}
		baseline.TargetProtocol = protocolList
		// Get telemetry parameter value with parameter type is 'fqdn'
		fqdnList, err := GetTelemetryParameterWithParameterTypeIsFqdn(engine, string(TELEMETRY_SETUP), vBaseline.Id, string(TARGET_FQDN))
		if err != nil {
			return nil, err
		}
		baseline.FQDN = fqdnList
		// Get telemetry parameter value with parameter type is 'uri'
		uriList, err := GetTelemetryParameterWithParameterTypeIsUri(engine, string(TELEMETRY_SETUP), vBaseline.Id, string(TARGET_URI))
		if err != nil {
			return nil, err
		}
		baseline.URI = uriList
		// Get telemetry parameter value with parameter type is 'alias-name'
		aliasNameList, err := GetTelemetryParameterWithParameterTypeIsAlias(engine, string(TELEMETRY_SETUP), vBaseline.Id, string(ALIAS_NAME))
		if err != nil {
			return nil, err
		}
		baseline.AliasName = aliasNameList
		// Get total traffic normal
		trafficList, err := GetTraffic(engine, string(TELEMETRY_SETUP), vBaseline.Id, string(TARGET_PREFIX), string(TOTAL_TRAFFIC_NORMAL))
		if err != nil {
			return nil, err
		}
		baseline.TotalTrafficNormal = trafficList
		// Get total traffic normal per protocol
		trafficPerProtocolList, err := GetTrafficPerProtocol(engine, string(TELEMETRY_SETUP), vBaseline.Id, string(TOTAL_TRAFFIC_NORMAL))
		if err != nil {
			return nil, err
		}
		baseline.TotalTrafficNormalPerProtocol = trafficPerProtocolList
		// Get total traffic normal per port
		trafficPerPortList, err := GetTrafficPerPort(engine, string(TELEMETRY_SETUP), vBaseline.Id, string(TOTAL_TRAFFIC_NORMAL))
		if err != nil {
			return nil, err
		}
		baseline.TotalTrafficNormalPerPort = trafficPerPortList
		// Get total connection capacity
		tccList, err := GetTotalConnectionCapacity(engine, vBaseline.Id)
		if err != nil {
			return nil, err
		}
		baseline.TotalConnectionCapacity = tccList
		// Get total connection capacity per port
		tccPerPortList, err := GetTotalConnectionCapacityPerPort(engine, vBaseline.Id)
		if err != nil {
			return nil, err
		}
		baseline.TotalConnectionCapacityPerPort = tccPerPortList
		// Get telemetry target list
		targetList, err := GetTelemetryTargetList(baseline.TargetPrefix, baseline.FQDN, baseline.URI)
		if err != nil {
			return nil, err
		}
		baseline.TargetList = targetList
		// Get alias data by alias name
		if len(baseline.AliasName) > 0 {
			aliasList, err := GetAliasByName(engine, customerId, cuid, baseline.AliasName.List())
			if err != nil {
				return nil, err
			}
			if len(aliasList.Alias) > 0 {
				aliasTargetList, err := GetAliasDataAsTargetList(aliasList)
				if err != nil {
					log.Errorf ("Failed to get alias data as target list. Error: %+v", err)
					return nil, err
				}
				// Append alias into target list
				baseline.TargetList = append(baseline.TargetList, aliasTargetList...)
			}
		}
		// Append baseline into baselineList
		baselineList = append(baselineList, baseline)
	}
	return
}

// Delete baseline
func DeleteBaseline(session *xorm.Session, id int64) (err error) {
	// Delete telemetry prefix
	err = db_models.DeleteTelemetryPrefix(session, string(TELEMETRY_SETUP), id, string(TARGET_PREFIX))
	if err != nil {
		log.Errorf("Delete telemetry prefix err: %+v", err)
		return
	}
	// Delete telemetry port range
	err = db_models.DeleteTelemetryPortRange(session, string(TELEMETRY_SETUP), id, string(TARGET_PREFIX))
	if err != nil {
		log.Errorf("Delete telemetry port range err: %+v", err)
		return
	}
	// Delete telemetry parameter values (protocol, fqdn, uri, alias-name)
	err = db_models.DeleteTelemetryParameterValue(session, string(TELEMETRY_SETUP), id)
	if err != nil {
		log.Errorf("Delete telemetry parameter value err: %+v", err)
		return
	}
	// Delete total traffic normal
	err = db_models.DeleteTraffic(session, string(TELEMETRY_SETUP), id, string(TARGET_PREFIX))
	if err != nil {
		log.Errorf("Delete telemetry traffic err: %+v", err)
		return
	}
	// Delete total traffic normal per protocol
	err = db_models.DeleteTrafficPerProtocol(session, string(TELEMETRY_SETUP), id)
	if err != nil {
		log.Errorf("Delete telemetry traffic per protocol err: %+v", err)
		return
	}
	// Delete total traffic normal per port
	err = db_models.DeleteTrafficPerPort(session, string(TELEMETRY_SETUP), id)
	if err != nil {
		log.Errorf("Delete telemetry traffic per port err: %+v", err)
		return
	}
	// Delete total connection capacity
	err = db_models.DeleteTotalConnectionCapacityByTeleBaselineId(session, id)
	if err != nil {
		log.Errorf("Delete total connection capacity err: %+v", err)
		return
	}
	// Delete total connection capacity per port
	err = db_models.DeleteTotalConnectionCapacityPerPortByTeleBaselineId(session, id)
	if err != nil {
		log.Errorf("Delete total connection capacity per port err: %+v", err)
		return
	}
	// Delete baseline
	err = db_models.DeleteBaselineById(session, id)
	if err != nil {
		log.Errorf("Delete baseline err: %+v", err)
		return
	}
	return
}

// Get telemetry prefix
func GetTelemetryPrefix(engine *xorm.Engine, tType string, typeId int64, prefixType string) (prefixList []Prefix, err error) {
	prefixs, err := db_models.GetTelemetryPrefix(engine, tType, typeId, prefixType)
	if err != nil {
		log.Errorf("Get telemetry prefix err: %+v", err)
		return nil, err
	}
	prefixList = []Prefix{}
	for _, vPrefix := range prefixs {
		prefix, err := NewPrefix(db_models.CreateIpAddress(vPrefix.Addr, vPrefix.PrefixLen))
		if err != nil {
			log.Errorf("Get telemetry prefix err: %+v", err)
			return nil, err
		}
		prefixList = append(prefixList, prefix)
	}
	return prefixList, nil
}

// Get telemetry port range
func GetTelemetryPortRange(engine *xorm.Engine, tType string, typeId int64, prefixType string) (portRangeList []PortRange, err error) {
	portRanges, err := db_models.GetTelemetryPortRange(engine, tType, typeId, prefixType)
	if err != nil {
		log.Errorf("Get telemetry port range err: %+v", err)
		return nil, err
	}
	portRangeList = []PortRange{}
	for _, vPortRange := range portRanges {
		portRange := PortRange{}
		portRange.LowerPort = vPortRange.LowerPort
		portRange.UpperPort = vPortRange.UpperPort
		portRangeList       = append(portRangeList, portRange)
	}
	return portRangeList, nil
}

// Get telemetry parameter with parameter type is 'protocol'
func GetTelemetryParameterWithParameterTypeIsProtocol(engine *xorm.Engine, tType string, typeId int64, parameterType string) (protocolList SetInt, err error) {
	protocolList = make(SetInt)
	protocols, err := db_models.GetTelemetryParameterValue(engine, tType, typeId, parameterType)
	if err != nil {
		log.Errorf("Get telemetry parameter with parameterType is 'protocol' err: %+v", err)
		return nil, err
	}
	for _, vProtocol := range protocols {
		protocolList.Append(vProtocol.IntValue)
	}
	return protocolList, nil
}

// Get telemetry parameter with parameter type is 'fqdn'
func GetTelemetryParameterWithParameterTypeIsFqdn(engine *xorm.Engine, tType string, typeId int64, parameterType string) (fqdnList SetString, err error) {
	fqdnList = make(SetString)
	fqdns, err := db_models.GetTelemetryParameterValue(engine, tType, typeId, parameterType)
	if err != nil {
		log.Errorf("Get telemetry parameter with parameterType is 'fqdn 'err: %+v", err)
		return nil, err
	}
	for _, vFqdn := range fqdns {
		fqdnList.Append(vFqdn.StringValue)
	}
	return fqdnList, nil
}

// Get telemetry parameter with parameter type is 'uri'
func GetTelemetryParameterWithParameterTypeIsUri(engine *xorm.Engine, tType string, typeId int64, parameterType string) (uriList SetString, err error) {
	uriList = make(SetString)
	uris, err := db_models.GetTelemetryParameterValue(engine, tType, typeId, parameterType)
	if err != nil {
		log.Errorf("Get telemetry parameter with parameterType is 'uri' err: %+v", err)
		return nil, err
	}
	for _, vUri := range uris {
		uriList.Append(vUri.StringValue)
	}
	return uriList, nil
}

// Get traffic
func GetTraffic(engine *xorm.Engine, tType string, typeId int64, prefixType string, trafficType string) (trafficList []Traffic, err error) {
	traffics, err := db_models.GetTraffic(engine, tType, typeId, prefixType, trafficType)
	if err != nil {
		log.Errorf("Get traffic err: %+v", err)
		return nil, err
	}
	trafficList = []Traffic{}
	for _, vTraffic := range traffics {
		traffic := Traffic{}
		traffic.Unit            = messages.ConvertUnitToInt(vTraffic.Unit)
		traffic.LowPercentileG  = messages.Uint64String(vTraffic.LowPercentileG)
		traffic.MidPercentileG  = messages.Uint64String(vTraffic.MidPercentileG)
		traffic.HighPercentileG = messages.Uint64String(vTraffic.HighPercentileG)
		traffic.PeakG           = messages.Uint64String(vTraffic.PeakG)
		trafficList             = append(trafficList, traffic)
	}
	return trafficList, nil
}

// Get traffic per protocol
func GetTrafficPerProtocol(engine *xorm.Engine, tType string, typeId int64, trafficType string) (trafficList []TrafficPerProtocol, err error) {
	traffics, err := db_models.GetTrafficPerProtocol(engine, tType, typeId, trafficType)
	if err != nil {
		log.Errorf("Get traffic per protocol err: %+v", err)
		return nil, err
	}
	trafficList = []TrafficPerProtocol{}
	for _, vTraffic := range traffics {
		traffic := TrafficPerProtocol{}
		traffic.Unit            = messages.ConvertUnitToInt(vTraffic.Unit)
		traffic.Protocol        = vTraffic.Protocol
		traffic.LowPercentileG  = messages.Uint64String(vTraffic.LowPercentileG)
		traffic.MidPercentileG  = messages.Uint64String(vTraffic.MidPercentileG)
		traffic.HighPercentileG = messages.Uint64String(vTraffic.HighPercentileG)
		traffic.PeakG           = messages.Uint64String(vTraffic.PeakG)
		trafficList             = append(trafficList, traffic)
	}
	return trafficList, nil
}

// Get traffic per port
func GetTrafficPerPort(engine *xorm.Engine, tType string, typeId int64, trafficType string) (trafficList []TrafficPerPort, err error) {
	traffics, err := db_models.GetTrafficPerPort(engine, tType, typeId, trafficType)
	if err != nil {
		log.Errorf("Get traffic per port err: %+v", err)
		return nil, err
	}
	trafficList = []TrafficPerPort{}
	for _, vTraffic := range traffics {
		traffic := TrafficPerPort{}
		traffic.Unit            = messages.ConvertUnitToInt(vTraffic.Unit)
		traffic.Port            = vTraffic.Port
		traffic.LowPercentileG  = messages.Uint64String(vTraffic.LowPercentileG)
		traffic.MidPercentileG  = messages.Uint64String(vTraffic.MidPercentileG)
		traffic.HighPercentileG = messages.Uint64String(vTraffic.HighPercentileG)
		traffic.PeakG           = messages.Uint64String(vTraffic.PeakG)
		trafficList             = append(trafficList, traffic)
	}
	return trafficList, nil
}

// Get total connection capacity
func GetTotalConnectionCapacity(engine *xorm.Engine, teleBaselineId int64) (tccList []TotalConnectionCapacity, err error) {
	tccs, err := db_models.GetTotalConnectionCapacityByTeleBaselineId(engine, teleBaselineId)
	if err != nil {
		log.Errorf("Get total connection capacity err: %+v", err)
		return nil, err
	}
	tccList = []TotalConnectionCapacity{}
	for _, vTcc := range tccs {
		tcc := TotalConnectionCapacity{}
		tcc.Protocol                = vTcc.Protocol
		tcc.Connection              = messages.Uint64String(vTcc.Connection)
		tcc.ConnectionClient        = messages.Uint64String(vTcc.ConnectionClient)
		tcc.Embryonic               = messages.Uint64String(vTcc.Embryonic)
		tcc.EmbryonicClient         = messages.Uint64String(vTcc.EmbryonicClient)
		tcc.ConnectionPs            = messages.Uint64String(vTcc.ConnectionPs)
		tcc.ConnectionClientPs      = messages.Uint64String(vTcc.ConnectionClientPs)
		tcc.RequestPs               = messages.Uint64String(vTcc.RequestPs)
		tcc.RequestClientPs         = messages.Uint64String(vTcc.RequestClientPs)
		tcc.PartialRequestMax       = messages.Uint64String(vTcc.PartialRequestMax)
		tcc.PartialRequestClientMax = messages.Uint64String(vTcc.PartialRequestClientMax)
		tccList                     = append(tccList, tcc)
	}
	return tccList, nil
}

// Get total connection capacity per port
func GetTotalConnectionCapacityPerPort(engine *xorm.Engine, teleBaselineId int64) (tccList []TotalConnectionCapacityPerPort, err error) {
	tccs, err := db_models.GetTotalConnectionCapacityPerPortByTeleBaselineId(engine, teleBaselineId)
	if err != nil {
		log.Errorf("Get total connection capacity per port err: %+v", err)
		return nil, err
	}
	tccList = []TotalConnectionCapacityPerPort{}
	for _, vTcc := range tccs {
		tcc := TotalConnectionCapacityPerPort{}
		tcc.Protocol                = vTcc.Protocol
		tcc.Port                    = vTcc.Port
		tcc.Connection              = messages.Uint64String(vTcc.Connection)
		tcc.ConnectionClient        = messages.Uint64String(vTcc.ConnectionClient)
		tcc.Embryonic               = messages.Uint64String(vTcc.Embryonic)
		tcc.EmbryonicClient         = messages.Uint64String(vTcc.EmbryonicClient)
		tcc.ConnectionPs            = messages.Uint64String(vTcc.ConnectionPs)
		tcc.ConnectionClientPs      = messages.Uint64String(vTcc.ConnectionClientPs)
		tcc.RequestPs               = messages.Uint64String(vTcc.RequestPs)
		tcc.RequestClientPs         = messages.Uint64String(vTcc.RequestClientPs)
		tcc.PartialRequestMax       = messages.Uint64String(vTcc.PartialRequestMax)
		tcc.PartialRequestClientMax = messages.Uint64String(vTcc.PartialRequestClientMax)
		tccList                     = append(tccList, tcc)
	}
	return tccList, nil
}

// Delete one telemetry setup configuration
func DeleteOneTelemetrySetup(customerId int, cuid string, cdid string, tsid int, dbTelemetrySetupList []db_models.TelemetrySetup) error {
	// database connection create
	engine, err := ConnectDB()
	if err != nil {
		log.Errorf("Database connect error: %s", err)
		return err
	}
	// transaction start
	session := engine.NewSession()
	defer session.Close()

	err = session.Begin()
	if err != nil {
		return err
	}
	for _, dbTelemetrySetup := range dbTelemetrySetupList {
		// Telemetry configuration
		if dbTelemetrySetup.SetupType == string(TELEMETRY_CONFIGURATION) {
			// Delete telemetry setup with setup type is 'telemetry_configuration'
			err = DeleteTelemetrySetupTelemetryConfiguration(engine, session, dbTelemetrySetup.Id)
			if err != nil {
				session.Rollback()
				return err
			}
			// Set default value for telemetry configuration
			telemetryConfiguration := DefaultValueTelemetryConfiguration()
			// Create telemetry setup with setup type is 'telemetry_configuration' 
			err = CreateTelemetrySetupTelemetryConfiguration(session, customerId, cuid, cdid, DefaultTsid, telemetryConfiguration)
			if err != nil {
				session.Rollback()
				return err
			}
		}

		// Pipe
		if  dbTelemetrySetup.SetupType == string(PIPE) {
			// Delete telemetry setup with setup type is 'pipe'
			err = DeleteTelemetrySetupPipe(engine, session, dbTelemetrySetup.Id)
			if err != nil {
				session.Rollback()
				return err
			}
			// Set default value for total_pipe_capacity
			pipeList := DefaultTotalPipeCapacity()
			// Create telemetry setup with setup type is 'pipe'
			err = CreateTelemetrySetupPipe(session, customerId, cuid, cdid, DefaultTsid, pipeList)
			if err != nil {
				session.Rollback()
				return err
			}
		}

		// Baseline
		if  dbTelemetrySetup.SetupType == string(BASELINE) {
			// Delete telemetry setup with setup type is 'baseline'
			err = DeleteTelemetrySetupBaseline(engine, session, customerId, cuid, dbTelemetrySetup.Id)
			if err != nil {
				session.Rollback()
				return err
			}
			// Set default value for baseline
			baselineList, err := DefaultBaseline()
			if err != nil {
				log.Errorf("Set default baseline err: %+v", err)
				return err
			}
			// Create telemetry setup with setup type is 'baseline'
			err = CreateTelemetrySetupBaseline(session, customerId, cuid, cdid, DefaultTsid, baselineList)
			if err != nil {
				session.Rollback()
				return err
			}
		}
	}
	// add Commit() after all actions
	err = session.Commit()
	return err
}

// Delete all telemetry configuration
func DeleteAllTelemetrySetup(customerId int, cuid string, cdid string) error {
	// database connection create
	engine, err := ConnectDB()
	if err != nil {
		log.Errorf("Database connect error: %s", err)
		return err
	}
	// transaction start
	session := engine.NewSession()
	defer session.Close()

	err = session.Begin()
	if err != nil {
		return err
	}
	// Get all telemetry configuration
	dbTelemetrySetupList, err := GetTelemetrySetupByCuid(customerId, cuid)
	if err != nil {
		log. Errorf("Get telemetry setup by cuid err: %+v", err)
		return err
	}
	for _, dbTelemetrySetup := range dbTelemetrySetupList {
		// Delete telemetry setup with setup type is 'telemetry_configuration'
		if dbTelemetrySetup.SetupType == string(TELEMETRY_CONFIGURATION) {
			err = DeleteTelemetrySetupTelemetryConfiguration(engine, session, dbTelemetrySetup.Id)
			if err != nil {
				session.Rollback()
				return err
			}
		}
		//Delete telemetry setup with setup type is 'pipe'
		if dbTelemetrySetup.SetupType == string(PIPE) {
			err = DeleteTelemetrySetupPipe(engine, session, dbTelemetrySetup.Id)
			if err != nil {
				session.Rollback()
				return err
			}
		}
		//telemetry setup with setup type is 'baseline'
		if dbTelemetrySetup.SetupType == string(BASELINE) {
			err = DeleteTelemetrySetupBaseline(engine, session, customerId, cuid, dbTelemetrySetup.Id)
			if err != nil {
				session.Rollback()
				return err
			}
		}
	}
	// Get default value for telemetry configuration
	telemetryConfiguration := DefaultValueTelemetryConfiguration()
	// Create telemetry setup with setup type is 'telemetry_configuration'
	err = CreateTelemetrySetupTelemetryConfiguration(session, customerId, cuid, cdid, DefaultTsid, telemetryConfiguration)
	if err != nil {
		session.Rollback()
		return err
	}
	// Get default value for total_pipe_capacity
	pipeList := DefaultTotalPipeCapacity()
	// Create telemetry setup with setup type is 'pipe'
	err = CreateTelemetrySetupPipe(session, customerId, cuid, cdid, DefaultTsid, pipeList)
	if err != nil {
		session.Rollback()
		return err
	}
	// Get default value for baseline
	baselineList, err := DefaultBaseline()
	if err != nil {
		log.Errorf("Get default baseline err: %+v", err)
		session.Rollback()
		return err
	}
	// Create telemetry setup with setup type is 'baseline'
	err = CreateTelemetrySetupBaseline(session, customerId, cuid, cdid, DefaultTsid, baselineList)
	if err != nil {
		session.Rollback()
		return err
	}
	// add Commit() after all actions
	err = session.Commit()
	return err
}

// Delete telemetry setup with setup type is 'telemetry_configuration'
func DeleteTelemetrySetupTelemetryConfiguration(engine *xorm.Engine, session *xorm.Session, teleSetupId int64) error {
	// Get telemetry configuration
	dbTelemetryConfiguration, err := db_models.GetTelemetryConfigurationByTeleSetupId(engine, teleSetupId)
	if err != nil {
		log.Errorf("Get telemetry configuration by teleSetupId err: %+v", err)
		return err
	}
	// Delete telemetry configuration
	if dbTelemetryConfiguration.Id > 0 {
		err = DeleteTelemetryConfiguration(session, dbTelemetryConfiguration.Id)
		if err != nil {
			return err
		}
	}
	// Delete telemetry setup
	err = db_models.DeleteTelemetrySetupById(session, teleSetupId)
	if err != nil {
		log.Errorf("Delete telemetry setup error: %s", err)
		return err
	}
	return nil
}

// Delete telemetry setup with setup type is 'pipe'
func DeleteTelemetrySetupPipe(engine *xorm.Engine, session *xorm.Session, teleSetupId int64) error {
	// Get total pipe capacity
	dbPipeList, err := db_models.GetTotalPipeCapacityByTeleSetupId(engine, teleSetupId)
	if err != nil {
		log.Errorf("Get total pipe capacity by teleSetupId err: %+v", err)
		session.Rollback()
		return err
	}
	// Delete total_pipe_capacity
	for _, dbPipe := range dbPipeList {
		err = db_models.DeleteTotalPipeCapacityById(session, dbPipe.Id)
		if err != nil {
			log.Errorf("Delete total pipe capacity err: %+v", err)
			session.Rollback()
			return err
		}
	}
	// Deleted telemetry setup
	err = db_models.DeleteTelemetrySetupById(session, teleSetupId)
	if err != nil {
		log.Errorf("Delete telemetry setup error: %s", err)
		session.Rollback()
		return err
	}
	return nil
}

// Delete telemetry setup with setup type is 'baseline'
func DeleteTelemetrySetupBaseline(engine *xorm.Engine, session *xorm.Session, customerId int, cuid string, teleSetupId int64) error {
	// Get baseline
	dbBaselineList, err := GetBaselineByTeleSetupId(customerId, cuid, teleSetupId)
	if err != nil {
		log.Errorf("Get baseline by teleSetupId err: %+v", err)
		session.Rollback()
		return err
	}
	// Delete baseline
	for _, dbBaseline := range dbBaselineList {
		err = DeleteBaseline(session, dbBaseline.Id)
		if err != nil {
			session.Rollback()
			return err
		}
	}
	// Deleted telemetry setup
	err = db_models.DeleteTelemetrySetupById(session, teleSetupId)
	if err != nil {
		log.Errorf("Delete telemetry setup error: %s", err)
		session.Rollback()
		return err
	}
	return nil
}

// Delete telemetry configuration
func DeleteTelemetryConfiguration(session *xorm.Session, teleConId int64) error {
    // Deleted unit configuration
	err := db_models.DeleteUnitConfigurationByTeleConfigId(session, teleConId)
	if err != nil {
		log.Errorf("Delete unit configuration err: %+v", err)
		return err
	}
	// Deleted telemetry configuration table data
	err = db_models.DeleteTelemetryConfigurationById(session, teleConId)
	if err != nil {
		log.Errorf("Delete telemetry configuration error: %s", err)
		return err
	}
	return nil
}

/*
 * Check overlap targets(target_prefix, target_fqdn, target_uri)
 * return:
 *    true: if request targets overlap with current targets
 *    false: else
 */
func CheckOverlapTargetList(requestTargets []Target, currentTargets []Target) bool {
	for _, requestTarget := range requestTargets {
		for _, currentTarget := range currentTargets {
			if requestTarget.TargetPrefix.Includes(&currentTarget.TargetPrefix) || currentTarget.TargetPrefix.Includes(&requestTarget.TargetPrefix) {
				log.Debugf("[Overlap] request target is: %+v with current target is: %+v", requestTarget.TargetPrefix, currentTarget.TargetPrefix)
				return true
			}
		}
	}
	return false
}

// Default value of telemetry configuration
func DefaultValueTelemetryConfiguration() (telemetryConfiguration *TelemetryConfiguration) {
	defaultValue := dots_config.GetServerSystemConfig().DefaultTelemetryConfiguration
	telemetryConfiguration = &TelemetryConfiguration{}
	telemetryConfiguration.MeasurementInterval       = messages.IntervalString(defaultValue.MeasurementInterval)
	telemetryConfiguration.MeasurementSample         = messages.SampleString(defaultValue.MeasurementSample)
	telemetryConfiguration.LowPercentile             = defaultValue.LowPercentile
	telemetryConfiguration.MidPercentile             = defaultValue.MidPercentile
	telemetryConfiguration.HighPercentile            = defaultValue.HighPercentile
	telemetryConfiguration.ServerOriginatedTelemetry = defaultValue.ServerOriginatedTelemetry
	telemetryConfiguration.TelemetryNotifyInterval   = defaultValue.TelemetryNotifyInterval

	unitConfig := UnitConfig{}
	unitConfig.Unit       = messages.UnitString(defaultValue.Unit)
	unitConfig.UnitStatus = defaultValue.UnitStatus
	telemetryConfiguration.UnitConfigList = append(telemetryConfiguration.UnitConfigList, unitConfig)
	return 
}

// Default value of total pipe capacity
func DefaultTotalPipeCapacity() (pipeList []TotalPipeCapacity) {
	defaultValue := dots_config.GetServerSystemConfig().DefaultTotalPipeCapacity
	pipeList = []TotalPipeCapacity{}
	pipe := TotalPipeCapacity{}
	pipe.LinkId   = defaultValue.LinkId
	pipe.Capacity = messages.Uint64String(defaultValue.Capacity)
	pipe.Unit     = messages.UnitString(defaultValue.Unit)
	pipeList = append(pipeList, pipe)
	return
}

// Default value of baseline
func DefaultBaseline() (baselineList []Baseline, err error) {
	defaultTargetValue := dots_config.GetServerSystemConfig().DefaultTarget
	defaultTtnbValue   := dots_config.GetServerSystemConfig().DefaultTotalTrafficNormalBaseline
	defaultTccValue    := dots_config.GetServerSystemConfig().DefaultTotalConnectionCapacity
	baselineList = []Baseline{}
	baseline := Baseline{}

	// target
	prefix, err := NewPrefix(defaultTargetValue.TargetPrefix)
	if err != nil {
		return nil, err
	}
	portRange := PortRange{}
	portRange.LowerPort = defaultTargetValue.TargetLowerPort
	portRange.UpperPort = defaultTargetValue.TargetUpperPort
	baseline.TargetPrefix    = append(baseline.TargetPrefix, prefix)
	baseline.TargetPortRange = append(baseline.TargetPortRange, portRange)
	baseline.TargetProtocol  = make(SetInt)
	baseline.FQDN            = make(SetString)
	baseline.URI             = make(SetString)
	baseline.TargetProtocol.Append(defaultTargetValue.TargetProtocol)
	baseline.FQDN.Append(defaultTargetValue.TargetFqdn)
	baseline.URI.Append(defaultTargetValue.TargetUri)

	// total_traffic_normal
	traffic := Traffic{}
	traffic.Unit            = messages.UnitString(defaultTtnbValue.Unit)
	traffic.LowPercentileG  = messages.Uint64String(defaultTtnbValue.LowPercentileG)
	traffic.MidPercentileG  = messages.Uint64String(defaultTtnbValue.MidPercentileG)
	traffic.HighPercentileG = messages.Uint64String(defaultTtnbValue.HighPercentileG)
	traffic.PeakG           = messages.Uint64String(defaultTtnbValue.PeakG)
	baseline.TotalTrafficNormal = append(baseline.TotalTrafficNormal, traffic)

	// total_connection_capacity
	tcc := TotalConnectionCapacity{}
	tcc.Protocol                = defaultTccValue.Protocol
	tcc.Connection              = messages.Uint64String(defaultTccValue.Connection)
	tcc.ConnectionClient        = messages.Uint64String(defaultTccValue.ConnectionClient)
	tcc.Embryonic               = messages.Uint64String(defaultTccValue.EmbryOnic)
	tcc.EmbryonicClient         = messages.Uint64String(defaultTccValue.EmbryOnicClient)
	tcc.ConnectionPs            = messages.Uint64String(defaultTccValue.ConnectionPs)
	tcc.ConnectionClientPs      = messages.Uint64String(defaultTccValue.ConnectionClientPs)
	tcc.RequestPs               = messages.Uint64String(defaultTccValue.RequestPs)
	tcc.RequestClientPs         = messages.Uint64String(defaultTccValue.RequestClientPs)
	tcc.PartialRequestMax       = messages.Uint64String(defaultTccValue.PartialRequestMax)
	tcc.PartialRequestClientMax = messages.Uint64String(defaultTccValue.PartialRequestClientMax)
	baseline.TotalConnectionCapacity = append(baseline.TotalConnectionCapacity, tcc)
	baselineList = append(baselineList, baseline)
	return baselineList, nil
}

// Get alias by name list (the name list from baseline/pre-mitigation)
func GetAliasByName(engine *xorm.Engine, customerId int, cuid string, aliasNames []string) (aliases types.Aliases, err error) {
	now := time.Now()
	aliases = types.Aliases{}
	client := db_models_data.Client{}
	// Get data client
	hasClient, err := engine.Where("customer_id=? AND cuid=?", customerId, cuid).Get(&client)
	if err != nil {
		log.Errorf("Failed to get client by cuid=%+v", cuid)
		return aliases, err
	}
	if !hasClient {
		return aliases, nil
	}
	for _, name := range aliasNames {
		if name == "" {
			continue
		}
		// Get data alias
		alias := db_models_data.Alias{}
		hasAlias, err := engine.Where("data_client_id = ? AND name = ?", client.Id, name).Get(&alias)
		if err != nil {
			log.Errorf("Failed to get alias by name=%+v", name)
			return aliases, err
		}
		if !hasAlias {
			log.Warnf("Alias with name: %+v has not been created by client: %+v", name, cuid)
			return aliases, nil
		} else if now.After(alias.ValidThrough) {
			continue
		}
		// Convert data alias to data types alias
		buf, err := json.Marshal(&alias.Alias)
		if err != nil {
			log.WithError(err).Error("ToTypesAlias - json.Marshal() failed.")
			return aliases, err
		}

		aliasTmp := types.Alias{}
		err = json.Unmarshal(buf, &aliasTmp)
		if err != nil {
			log.WithError(err).Error("ToTypesAlias - json.Unmarshal() failed.")
			return aliases, err
		}
		aliases.Alias = append(aliases.Alias, aliasTmp)
	}
	return aliases, nil
}
