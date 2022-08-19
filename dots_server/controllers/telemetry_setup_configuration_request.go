package controllers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"github.com/shopspring/decimal"
	"github.com/nttdots/go-dots/dots_common/messages"
	"github.com/nttdots/go-dots/dots_server/db_models"
	"github.com/nttdots/go-dots/dots_server/models"
	log "github.com/sirupsen/logrus"
	common "github.com/nttdots/go-dots/dots_common"
	types "github.com/nttdots/go-dots/dots_common/types/data"
	dots_config "github.com/nttdots/go-dots/dots_server/config"
	data_controllers "github.com/nttdots/go-dots/dots_server/controllers/data"
)

/*
 * Controller for the telemetryRequest API.
 */
 type TelemetrySetupRequest struct {
	Controller
}
/*
 * Handles telemetry PUT request
 *  1. The PUT telemetry_configuration
 *  2. The PUT total_pipe_capacity
 *  3. The PUT basline
 *
 * parameter:
 *  request request message
 *  customer request source Customer
 * return:
 *  res response message
 *  err error
 */
func (t *TelemetrySetupRequest) HandlePut(request Request, customer *models.Customer) (res Response, err error) {

	log.WithField("request", request).Debug("HandlePut")
	var errMsg string
	// Check Uri-Path cuid, tsid for telemetry setup request
	cuid, tsid, cdid, err := parseTelemetrySetupUriPath(request.PathInfo)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to parse Uri-Path, error: %s", err)
		log.Error(errMsg)
		res = Response {
			Type: common.Acknowledgement,
			Code: common.BadRequest,
			Body: errMsg,
		}
		return res, nil
	}
	if cuid == "" || tsid == nil {
		errMsg = "Missing required Uri-Path Parameter(cuid, tsid)."
		log.Error(errMsg)
		res = Response {
			Type: common.Acknowledgement,
			Code: common.BadRequest,
			Body: errMsg,
		}
		return res, nil
	}

	if request.Body == nil {
		errMsg = "Request body must be provided for PUT method"
		log.Error(errMsg)
		res = Response {
			Type: common.Acknowledgement,
			Code: common.BadRequest,
			Body: errMsg,
		}
		return res, nil
	}

	body := request.Body.(*messages.TelemetrySetupRequest)
	if len(body.TelemetrySetup.Telemetry) != 1 {
		// Zero or multiple telemetry setup configuration
		errMsg = "Request body must contain only one telemetry setup configuration"
		log.Error(errMsg)
		res = Response {
			Type: common.Acknowledgement,
			Code: common.BadRequest,
			Body: errMsg,
		}
		return res, nil
	}
	telemetry := body.TelemetrySetup.Telemetry[0]
	// A DOTS telemetry setup message MUST include only telememetry_configuration, total_pipe_capacity or baseline
	if (telemetry.TelemetryConfigurationCurrent != nil && len(telemetry.TotalPipeCapacity) > 0) ||
		(telemetry.TelemetryConfigurationCurrent != nil && len(telemetry.Baseline) > 0) ||
		(len(telemetry.Baseline) > 0 && len(telemetry.TotalPipeCapacity) > 0) {
		errMsg = "A DOTS telemetry setup message MUST include only telemetry-related configuration parameters or information about DOTS client domain pipe capacity or telemetry traffic baseline"
		log.Error(errMsg)
		res = Response {
			Type: common.Acknowledgement,
			Code: common.BadRequest,
			Body: errMsg,
		}
		return res, nil
	}
	if telemetry.TelemetryConfigurationCurrent != nil {
		// Handle Put telemetry_configuration
		res, err = handlePutTelemetryConfiguration(telemetry.TelemetryConfigurationCurrent, customer, cuid, cdid, *tsid)
	} else if len(telemetry.TotalPipeCapacity) > 0 {
		// Handle Put total_pipe_capacity
		res, err = handlePutTotalPipeCapacity(telemetry.TotalPipeCapacity, customer, cuid, cdid, *tsid)
	} else if len(telemetry.Baseline) > 0 {
		// Handle Put baseline
		res, err = handlePutBaseline(telemetry.Baseline, customer, cuid, cdid, *tsid)
	} else {
		// The body request doesn't contain telemetry setup data
		errMsg = "The body request doesn't contain telemetry setup data"
		log.Error(errMsg)
		res = Response {
			Type: common.Acknowledgement,
			Code: common.BadRequest,
			Body: errMsg,
		}
	}
	return res, err
}

/*
 * Handles telemetry GET request
 *  1. The Get all telemetry setup configuration when the uri-path doesn't contain 'tsid'
 *  2. The Get one telemetry setup configuration when the uri-path contains 'tsid'
 *
 * parameter:
 *  request request message
 *  customer request source Customer
 * return:
 *  res response message
 *  err error
 *
 */
func (t *TelemetrySetupRequest) HandleGet(request Request, customer *models.Customer) (res Response, err error) {
	log.WithField("request", request).Debug("[GET] receive message")
	var errMsg string
	// Check Uri-Path cuid, tsid for telemetry configuration request
	cuid, tsid, _, err := parseTelemetrySetupUriPath(request.PathInfo)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to parse Uri-Path, error: %s", err)
		log.Error(errMsg)
		res = Response {
			Type: common.Acknowledgement,
			Code: common.BadRequest,
			Body: errMsg,
		}
		return res, nil
	}
	if cuid == "" {
		errMsg = "Missing required Uri-Path Parameter cuid."
		log.Error(errMsg)
		res = Response {
			Type: common.Acknowledgement,
			Code: common.BadRequest,
			Body: errMsg,
		}
		return res, nil
	}
	if tsid != nil {
		log.Debug("Get one telemetry setup configuration")
		res, err = handleGetOneTelemetrySetup(customer.Id, cuid, *tsid)
	} else {
		log.Debug("Get all telemetry configuration")
		res, err = handleGetAllTelemetrySetup(customer.Id, cuid)
	}
	return
}

/*
 * Handles telemetry DELETE request
 *  1. The Delete all telemetry setup configuration when the uri-path doesn't contain 'tsid'
 *  2. The Delete one telemetry setup configuration when the uri-path contains 'tsid'
 *
 * parameter:
 *  request request message
 *  customer request source Customer
 * return:
 *  res response message
 *  err error
 *
 */
 func (t *TelemetrySetupRequest) HandleDelete(request Request, customer *models.Customer) (res Response, err error) {
	log.WithField("request", request).Debug("[DELETE] receive message")
	var errMsg string
	// Check Uri-Path cuid, tsid for telemetry configuration request
	cuid, tsid, cdid, err := parseTelemetrySetupUriPath(request.PathInfo)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to parse Uri-Path, error: %s", err)
		log.Error(errMsg)
		res = Response {
			Type: common.Acknowledgement,
			Code: common.BadRequest,
			Body: errMsg,
		}
		return res, nil
	}
	if cuid == "" {
		errMsg = "Missing required Uri-Path Parameter cuid."
		log.Error(errMsg)
		res = Response {
			Type: common.Acknowledgement,
			Code: common.BadRequest,
			Body: errMsg,
		}
		return res, nil
	}
	if tsid != nil {
		log.Debug("Delete one telemetry setup configuration")
		err = handleDeleteOneTelemetrySetup(customer.Id, cuid, cdid, *tsid)
		if err != nil {
			return Response{}, err
		}
	} else {
		log.Debug("Delete all telemetry setup configuration")
		err = models.DeleteAllTelemetrySetup(customer.Id, cuid, cdid)
		if err != nil {
			return Response{}, err
		}
	}
	res = Response{
		Type: common.Acknowledgement,
		Code: common.Deleted,
		Body: "Deleted",
	}
	return res, err
}

// Handle Put telemetry configuration
func handlePutTelemetryConfiguration(bodyRequest *messages.TelemetryConfigurationCurrent, customer *models.Customer, cuid string, cdid string, tsid int) (res Response, err error) {
	// validate TelemetryConfiguration
	isPresent, isUnprocessableEntity, errMsg := models.ValidateTelemetryConfiguration(customer.Id, cuid, tsid, bodyRequest)
	if errMsg != "" {
		if isUnprocessableEntity {
			res = Response {
				Type: common.Acknowledgement,
				Code: common.UnprocessableEntity,
				Body: errMsg,
			}
			return res, nil
		}
		res = Response {
			Type: common.Acknowledgement,
			Code: common.BadRequest,
			Body: errMsg,
		}
		return res, nil
	}
	telemetryConfiguration := models.NewTelemetryConfiguration(bodyRequest)
	// If 'tsid' doesn't exist in DB, the DOTS server will create new telemetry configuration
	// Else the DOTS server will update telemetry configuration
	if !isPresent {
		log.Debug("Create telemetry configuration")
		err := models.CreateTelemetryConfiguration(customer.Id, cuid, cdid, tsid, telemetryConfiguration)
		if err != nil {
			return Response{}, err
		}
		res = Response{
			Type: common.Acknowledgement,
			Code: common.Created,
			Body: nil,
		}
		return res, nil
	}
	log.Debug("Update telemetry configuration")
	err = models.UpdateTelemetryConfiguration(customer.Id, cuid, cdid, tsid, telemetryConfiguration)
	if err != nil {
		return Response{}, err
	}
	res = Response{
		Type: common.Acknowledgement,
		Code: common.Changed,
		Body: nil,
	}
	return res, nil
}

// Handle Put total pipe capacity
func handlePutTotalPipeCapacity(bodyRequest []messages.TotalPipeCapacity, customer *models.Customer, cuid string, cdid string, tsid int) (res Response, err error) {
	var conflictInfo *models.ConflictInformation
	// validate TotalPipeCapacity
	isPresent, isUnprocessableEntity, errMsg := models.ValidateTotalPipeCapacity(customer.Id, cuid, tsid, bodyRequest)
	if errMsg != "" {
		if isUnprocessableEntity {
			res = Response {
				Type: common.Acknowledgement,
				Code: common.UnprocessableEntity,
				Body: errMsg,
			}
			return res, nil
		}
		res = Response {
			Type: common.Acknowledgement,
			Code: common.BadRequest,
			Body: errMsg,
		}
		return res, nil
	}
	totalPipeCapacity := models.NewTotalPipeCapacity(bodyRequest)
	// If 'tsid' doesn't exist in DB, the DOTS server will create new total pipe capacity
	// Else the DOTS server will update total pipe capacity
	isConflict, err := models.CreateTotalPipeCapacity(customer.Id, cuid, cdid, tsid, totalPipeCapacity, isPresent)
	if err != nil {
		return Response{}, err
	}
	// Created conflict information with conflict cause is 'overlap_targets'
	if isConflict {
		log.Error("[Conflicted] Existed total pipe capacity")
		conflictInfo = &models.ConflictInformation {
			ConflictCause:  models.OVERLAPPING_PIPE_SCOPE,
			ConflictScope:  nil,
		}
		res = Response {
			Type: common.Acknowledgement,
			Code: common.Conflict,
			Body: messages.NewTelemetrySetupConfigurationResponseConflict(tsid, conflictInfo.ParseToResponse()),
		}
		return res, nil
	}
	if !isPresent {
		res = Response{
			Type: common.Acknowledgement,
			Code: common.Created,
			Body: nil,
		}
		return res, nil
	}
	res = Response{
		Type: common.Acknowledgement,
		Code: common.Changed,
		Body: nil,
	}
	return res, nil
}

// Handle Put baseline
func handlePutBaseline(bodyRequest []messages.Baseline, customer *models.Customer, cuid string, cdid string, tsid int) (res Response, err error) {
	var conflictInfo *models.ConflictInformation
	// validate baseline
	isPresent, isUnprocessableEntity, errMsg := models.ValidateBaseline(customer, cuid, tsid, bodyRequest)
	if errMsg != "" {
		if isUnprocessableEntity {
			res = Response {
				Type: common.Acknowledgement,
				Code: common.UnprocessableEntity,
				Body: errMsg,
			}
			return res, nil
		}
		res = Response {
			Type: common.Acknowledgement,
			Code: common.BadRequest,
			Body: errMsg,
		}
		return res, nil
	}
	// Get data alias from data channel
	var aliases types.Aliases
	if bodyRequest[0].AliasName != nil {
		aliases, err = data_controllers.GetDataAliasesByName(customer, cuid, bodyRequest[0].AliasName)
		if err != nil {
			log.Errorf("Get data alias error: %+v", err)
			return Response{}, err
		}
		if len(aliases.Alias) <= 0 {
			errMsg = "'alias-name' doesn't exist in DB"
			res = Response {
				Type: common.Acknowledgement,
				Code: common.NotFound,
				Body: errMsg,
			}
			return res, nil
		}
	}
	baselineList, err := models.NewBaselineList(bodyRequest, aliases)
	if err != nil {
		return Response{}, err
	}
	// If 'tsid' doesn't exist in DB, the DOTS server will create new baseline
	// Else the DOTS server will update baseline
	isConflict, err := models.CreateBaseline(customer.Id, cuid, cdid, tsid, baselineList, isPresent)
	if err != nil {
		return Response{}, err
	}
	if isConflict {
		log.Error("[Conflicted] Existed baseline")
		conflictInfo = &models.ConflictInformation {
			ConflictCause:  models.OVERLAPPING_TARGETS,
			ConflictScope:  nil,
		}
		res = Response {
			Type: common.Acknowledgement,
			Code: common.Conflict,
			Body: messages.NewTelemetrySetupConfigurationResponseConflict(tsid, conflictInfo.ParseToResponse()),
		}
		return res, nil
	}
	if !isPresent {
		res = Response{
			Type: common.Acknowledgement,
			Code: common.Created,
			Body: nil,
		}
		return res, nil
	}
	res = Response{
		Type: common.Acknowledgement,
		Code: common.Changed,
		Body: nil,
	}
	return res, nil
}

// Handle Get one telemetry setup configuration
func handleGetOneTelemetrySetup(customerId int, cuid string, tsid int) (res Response, err error) {
	// Get telemetry setup configuration by 'tsid' from DB
	// The telemetry setup configuration with setup_type is 'telemetry_configuration', 'pipe' and 'baseline'
	dbTelemetrySetupList, err := models.GetTelemetrySetupByTsid(customerId, cuid, tsid)
	if err != nil {
		log. Errorf("Get telemetry_setup by tsid err: %+v", err)
		return Response{}, err
	}
	// If 'tsid' doesn't exist in DB, DOTS server will response 4.04 NotFound
	// Else Get telemetry_configuration, total_pipe_capacity or baseline with 'tsid' existed in DB
	if len(dbTelemetrySetupList) < 1 {
		errMsg := fmt.Sprintf("Not found telemetry setup with tsid = %+v", tsid)
		log.Error(errMsg)
		res = Response{
			Type: common.Acknowledgement,
			Code: common.NotFound,
			Body: errMsg,
		}
		return res, nil
	}
	telemetrySetupResp := messages.TelemetrySetupResponse{}
	// Get telemetry setup
	telemetry := messages.TelemetryResponse{}
	telemetry.Tsid = tsid
	for _, dbTelemetrySetup := range dbTelemetrySetupList {
		err = getTelemetrySetup(customerId, cuid, dbTelemetrySetup, &telemetry)
		if err != nil {
			return Response{}, err
		}
	}
	telemetrySetupResp.TelemetrySetup.Telemetry = append(telemetrySetupResp.TelemetrySetup.Telemetry, telemetry)
	res = Response{
		Type: common.Acknowledgement,
		Code: common.Content,
		Body: telemetrySetupResp,
	}
	return res, nil
}

// Handle Get all telemetry configuration
func handleGetAllTelemetrySetup(customerId int, cuid string) (res Response, err error) {
	telemetrySetupResp := messages.TelemetrySetupResponse{}
	// Get telemetry setup configuration by 'cuid' from DB
	// The telemetry setup configuration with setup_type is 'telemetry_configuration', 'pipe' and 'baseline'
	dbTelemetrySetupList, err := models.GetTelemetrySetupByCuidAndNonNegativeTsid(customerId, cuid)
	if err != nil {
		log. Errorf("Get telemetry_setup by cuid err: %+v", err)
		return Response{}, err
	}
	if len(dbTelemetrySetupList) > 0 {
		telemetrySetupList := messages.TelemetrySetupResp{}
		for _, vDbTelemetrySetup := range dbTelemetrySetupList {
			if vDbTelemetrySetup.Tsid > 0 {
				telemetry := messages.TelemetryResponse{}
				telemetry.Tsid = vDbTelemetrySetup.Tsid
				for k, vTelemetrySetup := range telemetrySetupList.Telemetry {
					// If 'tsid' same with 'tsid' of TelemetrySetupResponse in TelemetrySetupResponseList, DOTS server will handle as below:
					// - Set new TelemetrySetupResponse is this TelemetrySetupResponse
					// - Removed this TelemetrySetupResponse in TelemetrySetupResponseList
					// - Set value of telemetry configuration, total pipe capacity or baseline with 'tsid' into new TelemetrySetupResponse
					// DOTS server will remove  telemetry setup in telemetry setup list
					if vDbTelemetrySetup.Tsid == vTelemetrySetup.Tsid {
						telemetry = vTelemetrySetup
						telemetrySetupList.Telemetry = append(telemetrySetupList.Telemetry[:k], telemetrySetupList.Telemetry[k+1:]...)
					}
				}
				// Get telemetry_configuration, total_pipe_capacity or baseline with 'cuid'
				err = getTelemetrySetup(customerId, cuid, vDbTelemetrySetup, &telemetry)
				if err != nil {
					return Response{}, err
				}
				telemetrySetupList.Telemetry = append(telemetrySetupList.Telemetry, telemetry)
			}
		}
		telemetrySetupResp.TelemetrySetup = telemetrySetupList
	}
	// Get min-max config values
	maxConfig, minConfig, supportedUnit := getMinMaxConfigValues()
	telemetrySetupResp.TelemetrySetup.MaxConfig     = maxConfig
	telemetrySetupResp.TelemetrySetup.MinConfig     = minConfig
	telemetrySetupResp.TelemetrySetup.SupportedUnit = supportedUnit
	// query-type config value
	telemetrySetupResp.TelemetrySetup.QueryType = dots_config.GetServerSystemConfig().QueryType
	res = Response{
		Type: common.Acknowledgement,
		Code: common.Content,
		Body: telemetrySetupResp,
	}
	return res, nil
}

// Handle delete one telemetry configuration
func handleDeleteOneTelemetrySetup(customerId int, cuid string, cdid string, tsid int) (err error) {
	// Get telemetry setup configuration by 'tsid' from DB
	dbTelemetrySetupList, err := models.GetTelemetrySetupByTsid(customerId, cuid, tsid)
	if err != nil {
		log. Errorf("Get telemetry_setup err: %+v", err)
		return err
	}
	// If 'tsid' exist/doesn't exist in DB, DOTS server will response 2.02 Deleted
	if len(dbTelemetrySetupList) > 0 {
		err = models.DeleteOneTelemetrySetup(customerId, cuid, cdid, tsid, dbTelemetrySetupList)
		if err != nil {
			return err
		}
	}
	return nil
}

// Get telemetry setup configuration contains setup_type 'telemetry_configuration', 'total_pipe_capacity' or 'baseline'
func getTelemetrySetup(customerId int, cuid string, dbTelemetrySetup db_models.TelemetrySetup,telemetry *messages.TelemetryResponse) (err error) {
	// Get telemetry configuration
	if dbTelemetrySetup.SetupType == string(models.TELEMETRY_CONFIGURATION) {
		currentConfig, err := getTelemetryConfiguration(dbTelemetrySetup.Id)
		if err != nil {
			return err
		}
		telemetry.CurrentConfig = currentConfig
	} else if dbTelemetrySetup.SetupType == string(models.PIPE) {
		// Get total pipe capapcity
		pipeList, err := getTotalPipeCapacity(dbTelemetrySetup.Id)
		if err != nil {
			return err
		}
		telemetry.TotalPipeCapacity = pipeList
	} else if dbTelemetrySetup.SetupType == string(models.BASELINE) {
		// Get baseline
		baselineList, err := getBaseline(customerId, cuid, dbTelemetrySetup.Id)
		if err != nil {
			return err
		}
		telemetry.Baseline = baselineList
	}
	return nil
}

// Get min-max config values
func getMinMaxConfigValues()(maxConfig *messages.TelemetryConfigurationResponse, minConfig *messages.TelemetryConfigurationResponse, supportedUnit *messages.SupportedUnitResponse) {
	maxConfig = &messages.TelemetryConfigurationResponse{}
	minConfig = &messages.TelemetryConfigurationResponse{}
	supportedUnit = &messages.SupportedUnitResponse{}
	// Get config value from config file
	config := dots_config.GetServerSystemConfig().TelemetryConfigurationParameter
	if config != nil {
		// Set Max of telemetry configuration from config value
		maxInterval := config.MeasurementInterval.End().(int)
		maxSample   := config.MeasurementSample.End().(int)
		maxConfig.MeasurementInterval       = messages.IntervalString(maxInterval)
		maxConfig.MeasurementSample         = messages.SampleString(maxSample)
		maxConfig.LowPercentile             = decimal.NewFromFloat(config.LowPercentile.End().(float64)).Round(2)
		maxConfig.MidPercentile             = decimal.NewFromFloat(config.MidPercentile.End().(float64)).Round(2)
		maxConfig.HighPercentile            = decimal.NewFromFloat(config.HighPercentile.End().(float64)).Round(2)
		maxConfig.ServerOriginatedTelemetry = &config.ServerOriginatedTelemetry
		maxTelemetryNotifyInterval          := config.TelemetryNotifyInterval.End().(int)
		maxConfig.TelemetryNotifyInterval   = &maxTelemetryNotifyInterval

		// Set Min of telemetry configuration from config value
		minInterval := config.MeasurementInterval.Start().(int)
		minSample   := config.MeasurementSample.Start().(int)
		minConfig.MeasurementInterval     = messages.IntervalString(minInterval)
		minConfig.MeasurementSample       = messages.SampleString(minSample)
		minConfig.LowPercentile           = decimal.NewFromFloat(config.LowPercentile.Start().(float64)).Round(2)
		minConfig.MidPercentile           = decimal.NewFromFloat(config.MidPercentile.Start().(float64)).Round(2)
		minConfig.HighPercentile          = decimal.NewFromFloat(config.HighPercentile.Start().(float64)).Round(2)
		minTelemetryNotifyInterval        := config.TelemetryNotifyInterval.Start().(int)
		minConfig.TelemetryNotifyInterval = &minTelemetryNotifyInterval

		// Set UnitConfig of telemetry configuration from config value
		unitConfig := messages.UnitConfigResponse{Unit: messages.UnitString(config.Unit), UnitStatus: config.UnitStatus}
		supportedUnit.UnitConfigList = append(supportedUnit.UnitConfigList, unitConfig)
	}
	return
}

// Get telemetry configuration
func getTelemetryConfiguration(dbTelemetrySetupId int64) (currentConfig *messages.TelemetryConfigurationResponse, err error) {
	currentConfig = &messages.TelemetryConfigurationResponse{}
	teleConfig := &models.TelemetryConfiguration{}

	// If telemetry setup with setup_type is 'telemetry_configuration' doesn't exist in DB, DOTS server will set value of telemetry configuration is default value
	// Else DOTS server will set value of telemetry configuration is value that is get from DB
	if dbTelemetrySetupId <= 0 {
		// Get default value for telemetry configuration
		teleConfig = models.DefaultValueTelemetryConfiguration()
	} else {
		// Get telemetry configuration
		teleConfig, err = models.GetTelemetryConfiguration(dbTelemetrySetupId)
		if err != nil {
			return  nil, err
		}
	}
	if teleConfig != nil {
		currentConfig.MeasurementInterval = messages.IntervalString(teleConfig.MeasurementInterval)
		currentConfig.MeasurementSample = messages.SampleString(teleConfig.MeasurementSample)
		currentConfig.LowPercentile = decimal.NewFromFloat(teleConfig.LowPercentile).Round(2)
		currentConfig.MidPercentile = decimal.NewFromFloat(teleConfig.MidPercentile).Round(2)
		currentConfig.HighPercentile = decimal.NewFromFloat(teleConfig.HighPercentile).Round(2)
		unitConfigList := []messages.UnitConfigResponse{}
		for _, v := range teleConfig.UnitConfigList {
			unitConfig := messages.UnitConfigResponse{Unit: v.Unit, UnitStatus: v.UnitStatus}
			unitConfigList = append(unitConfigList, unitConfig)
		}
		currentConfig.UnitConfigList = unitConfigList
		currentConfig.ServerOriginatedTelemetry = &teleConfig.ServerOriginatedTelemetry
		if teleConfig.TelemetryNotifyInterval >= 1 {
			currentConfig.TelemetryNotifyInterval = &teleConfig.TelemetryNotifyInterval
		}
	}
	return currentConfig, nil
}

// Get total pipe capacity
func getTotalPipeCapacity(dbTelemetrySetupId int64) (pipeList []messages.TotalPipeCapacityResponse, err error) {
	pipeList = []messages.TotalPipeCapacityResponse{}
	totalPipeCapacityList := []models.TotalPipeCapacity{}
	// If telemetry setup with setup_type is 'pipe' doesn't exist in DB, DOTS server will set value of total pipe capacity is default value
	// Else DOTS server will set value of total pipe capacity is value that is get from DB
	if dbTelemetrySetupId <= 0 {
		// Get default value for total_pipe_capacity
		totalPipeCapacityList = models.DefaultTotalPipeCapacity()
	} else {
		// Get total pipe capacity
		totalPipeCapacityList, err = models.GetTotalPipeCapacity(dbTelemetrySetupId)
		if err != nil {
			return nil, err
		}
	}
	for _, v := range totalPipeCapacityList {
		totalPipecapacity := messages.TotalPipeCapacityResponse{}
		totalPipecapacity.LinkId   = v.LinkId
		totalPipecapacity.Capacity = v.Capacity
		totalPipecapacity.Unit     = v.Unit
		pipeList                   = append(pipeList, totalPipecapacity)
	}
	return pipeList, nil
}

// Get baseline
func getBaseline(customerId int, cuid string, teleSetupId int64) (baselineList []messages.BaselineResponse, err error) {
	baselineList = []messages.BaselineResponse{}
	baselines := []models.Baseline{}
	// If telemetry setup with setup_type is 'baseline' doesn't exist in DB, DOTS server will set value of baseline is default value
	// Else DOTS server will set value of baseline is value that is get from DB
	if teleSetupId <= 0 {
		// Get default value for baseline
		baselines, err = models.DefaultBaseline()
		if err != nil {
			log.Errorf("Get default baseline err: %+v", err)
			return nil, err
		}
	} else {
		// Get baseline
		baselines, err = models.GetBaselineByTeleSetupId(customerId, cuid, teleSetupId)
		if err != nil {
			return nil, err
		}
	}
	for _, v := range baselines {
		baseline := messages.BaselineResponse{}
		baseline.Id = v.BaselineId
		// target
		for _, vPrefix := range v.TargetPrefix {
			baseline.TargetPrefix = append(baseline.TargetPrefix, vPrefix.String())
		}
		for _, vPortRange := range v.TargetPortRange {
			baseline.TargetPortRange = append(baseline.TargetPortRange, messages.PortRangeResponse{LowerPort: vPortRange.LowerPort, UpperPort: &vPortRange.UpperPort})
		}
		for _, vProtocol := range v.TargetProtocol.List() {
			baseline.TargetProtocol = append(baseline.TargetProtocol, vProtocol)
		}
		for _, vFqdn := range v.FQDN.List() {
			baseline.TargetFQDN = append(baseline.TargetFQDN, vFqdn)
		}
		for _, vUri := range v.URI.List() {
			baseline.TargetURI = append(baseline.TargetURI, vUri)
		}
		for _, vAliasName := range v.AliasName.List() {
			baseline.AliasName = append(baseline.AliasName, vAliasName)
		}
		// total traffic normal
		baseline.TotalTrafficNormal = convertToTrafficResponse(v.TotalTrafficNormal)
		// total traffic normal per protocol
		baseline.TotalTrafficNormalPerProtocol = convertToTrafficPerProtocolResponse(v.TotalTrafficNormalPerProtocol)
		// total traffic normal per port
		baseline.TotalTrafficNormalPerPort = convertToTrafficPerPortResponse(v.TotalTrafficNormalPerPort)
		// total connection capacity
		baseline.TotalConnectionCapacity = convertToTotalConnectionCapacityResponse(v.TotalConnectionCapacity)
		// total connection capacity per port
		baseline.TotalConnectionCapacityPerPort = convertToTotalConnectionCapacityPerPortResponse(v.TotalConnectionCapacityPerPort)
		baselineList = append(baselineList, baseline)
	}
	return baselineList, nil
}

/*
 *  Get cuid, tsid, cdid value from URI-Path
 */
 func parseTelemetrySetupUriPath(uriPath []string) (cuid string, tsid *int, cdid string, err error){
	log.Debugf("Parsing URI-Path : %+v", uriPath)
	// Get cuid, cdid, tsid from Uri-Path
	// If Uri-Path contains one or more invalid or unknown parameter, DOTS server will response 400 Bad Request
	for _, uriPath := range uriPath{
		if(strings.HasPrefix(uriPath, "cuid=")){
			cuid = uriPath[strings.Index(uriPath, "cuid=")+5:]
		} else if(strings.HasPrefix(uriPath, "cdid=")){
			cuid = uriPath[strings.Index(uriPath, "cdid=")+5:]
		} else if(strings.HasPrefix(uriPath, "tsid=")){
			tcidStr := uriPath[strings.Index(uriPath, "tsid=")+5:]
			tcidValue, err := strconv.Atoi(tcidStr)
			if err != nil {
				log.Error("Tsid is not integer type.")
				return cuid, tsid, cdid, err
			}
			if tcidStr == "" {
			    tsid = nil
			} else {
			    tsid = &tcidValue
			}
		} else if !(strings.HasPrefix(uriPath, "well-known")) && !(strings.HasPrefix(uriPath, "dots")) && !(strings.HasPrefix(uriPath, "tm-setup")) {
			err = errors.New("Uri-Path MUST NOT contains one or more invalid or unknown parameters")
			log.Error(err)
			return cuid, tsid, cdid, err
		}
	}
	// Log nil if tsid does not exist in path. Otherwise, log tsid's value
	if tsid == nil {
	    log.Debugf("Parsing URI-Path result : cuid=%+v, tsid=%+v", cuid, nil)
	} else {
        log.Debugf("Parsing URI-Path result : cuid=%+v, tsid=%+v", cuid, *tsid)
	}
	return
}