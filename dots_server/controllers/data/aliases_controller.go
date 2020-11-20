package data_controllers

import (
  "net/http"
  "time"
  "fmt"

  "github.com/julienschmidt/httprouter"
  log "github.com/sirupsen/logrus"

  messages_common "github.com/nttdots/go-dots/dots_common/messages"
  messages "github.com/nttdots/go-dots/dots_common/messages/data"
  types    "github.com/nttdots/go-dots/dots_common/types/data"
  "github.com/nttdots/go-dots/dots_server/db"
  "github.com/nttdots/go-dots/dots_server/models"
  "github.com/nttdots/go-dots/dots_server/models/data"
)

const (
  DEFAULT_ALIAS_LIFETIME_IN_MINUTES = 7 * 1440
)

var defaultAliasLifetime = DEFAULT_ALIAS_LIFETIME_IN_MINUTES * time.Minute

type AliasesController struct {
}

func (c *AliasesController) GetAll(customer *models.Customer, r *http.Request, p httprouter.Params) (Response, error) {
  now := time.Now()
  cuid := p.ByName("cuid")
  log.WithField("cuid", cuid).Info("[AliasesController] GET")
  isAfterTransaction := false

  // Check missing 'cuid'
  if cuid == "" {
    log.Error("Missing required path 'cuid' value.")
    return ErrorResponse(http.StatusBadRequest, ErrorTag_Missing_Attribute, "Missing a mandatory attribute : 'cuid'", isAfterTransaction)
  }

  return WithTransaction(func (tx *db.Tx) (Response, error) {
    return WithClient(tx, customer, cuid, func (client *data_models.Client) (_ Response, err error) {
      aliases, err := data_models.FindAliases(tx, client, now)
      if err != nil {
        return
      }

      tas, err := aliases.ToTypesAliases(now)
      if err != nil {
        return
      }
      s := messages.AliasesResponse{}
      s.Aliases = *tas

      cp, err := messages.ContentFromRequest(r)
      if err != nil {
        return
      }

      m, err := messages.ToMap(s, cp)
      if err != nil {
        return
      }
      return YangJsonResponse(m)
    })
  })
}

func (c *AliasesController) Get(customer *models.Customer, r *http.Request, p httprouter.Params) (Response, error) {
  now := time.Now()
  cuid := p.ByName("cuid")
  name := p.ByName("alias")
  log.WithField("cuid", cuid).WithField("alias", name).Info("[AliasesController] GET")
  isAfterTransaction := false

  // Check missing 'cuid'
  if cuid == "" {
    log.Error("Missing required path 'cuid' value.")
    return ErrorResponse(http.StatusBadRequest, ErrorTag_Missing_Attribute, "Missing a mandatory attribute : 'cuid'", isAfterTransaction)
  }

  // Check missing alias 'name'
  if name == "" {
    log.Error("Missing required alias 'name' attribute.")
    return ErrorResponse(http.StatusBadRequest, ErrorTag_Missing_Attribute, "Missing a mandatory attribute : alias 'name'", isAfterTransaction)
  }

  return WithTransaction(func (tx *db.Tx) (Response, error) {
    return WithClient(tx, customer, cuid, func (client *data_models.Client) (_ Response, err error) {
      isAfterTransaction = true
      alias, err := data_models.FindAliasByName(tx, client, name, now)
      if err != nil {
        return
      }
      if alias == nil {
        errMsg := fmt.Sprintf("Not Found alias by specified name = %+v", name)
        return ErrorResponse(http.StatusNotFound, ErrorTag_Invalid_Value, errMsg, isAfterTransaction)
      }

      ta, err := alias.ToTypesAlias(now)
      if err != nil {
        return
      }
      s := messages.AliasesResponse{}
      s.Aliases.Alias = []types.Alias{ *ta }

      cp, err := messages.ContentFromRequest(r)
      if err != nil {
        return
      }

      m, err := messages.ToMap(s, cp)
      if err != nil {
        return
      }
      return YangJsonResponse(m)
    })
  })
}

func (c *AliasesController) Delete(customer *models.Customer, r *http.Request, p httprouter.Params) (Response, error) {
  now := time.Now()
  cuid := p.ByName("cuid")
  name := p.ByName("alias")
  log.WithField("cuid", cuid).WithField("alias", name).Info("[AliasesController] DELETE")
  isAfterTransaction := false

   // Check missing 'cuid'
   if cuid == "" {
    log.Error("Missing required path 'cuid' value.")
    return ErrorResponse(http.StatusBadRequest, ErrorTag_Missing_Attribute, "Missing a mandatory attribute : 'cuid'", isAfterTransaction)
  }

  // Check missing alias 'name'
  if name == "" {
    log.Error("Missing required alias 'name' attribute.")
    return ErrorResponse(http.StatusBadRequest, ErrorTag_Missing_Attribute, "Missing a mandatory attribute : alias 'name'", isAfterTransaction)
  }

  return WithTransaction(func (tx *db.Tx) (Response, error) {
    return WithClient(tx, customer, cuid, func (client *data_models.Client) (_ Response, err error) {
      isAfterTransaction = true
      deleted, err := data_models.DeleteAliasByName(tx, client.Id, name, now)
      if err != nil {
        return
      }

      if deleted == true {
        return EmptyResponse(http.StatusNoContent)
      } else {
        errMsg := fmt.Sprintf("Not Found alias by specified name = %+v", name)
        return ErrorResponse(http.StatusNotFound, ErrorTag_Invalid_Value, errMsg, isAfterTransaction)
      }
    })
  })
}

func (c *AliasesController) Put(customer *models.Customer, r *http.Request, p httprouter.Params) (Response, error) {
  now := time.Now()
  cuid := p.ByName("cuid")
  cdid := p.ByName("cdid")
  name := p.ByName("alias")
  errMsg := ""
  log.WithField("cuid", cuid).WithField("alias", name).Info("[AliasesController] PUT")
  isAfterTransaction := false

  // Check missing 'cuid'
  if cuid == "" {
    log.Error("Missing required path 'cuid' value.")
    return ErrorResponse(http.StatusBadRequest, ErrorTag_Missing_Attribute, "Missing a mandatory attribute : 'cuid'", isAfterTransaction)
  }

  // Check missing alias 'name'
  if name == "" {
    log.Error("Missing required path 'name' value.")
    return ErrorResponse(http.StatusBadRequest, ErrorTag_Missing_Attribute, "Missing a mandatory attribute : alias 'name'", isAfterTransaction)
  }

  req := messages.AliasesRequest{}
  err := Unmarshal(r, &req)
  if err != nil {
    errMsg = fmt.Sprintf("Invalid body data format: %+v", err)
    return ErrorResponse(http.StatusBadRequest, ErrorTag_Invalid_Value, errMsg, isAfterTransaction)
  }
  log.Infof("[AliasesController] Put request=%#+v", req)

  // Get blocker configuration by customerId and target_type in table blocker_configuration
	blockerConfig, err := models.GetBlockerConfiguration(customer.Id, string(messages_common.DATACHANNEL_ACL))
	if err != nil {
		return ErrorResponse(http.StatusInternalServerError, ErrorTag_Operation_Failed, "Get blocker configuration failed", isAfterTransaction)
	}
	log.WithFields(log.Fields{
		"blocker_type": blockerConfig.BlockerType,
  }).Debug("Get blocker configuration")

  // Validation
  validator := messages.GetAliasValidator(blockerConfig.BlockerType)
  if validator == nil {
    errString := fmt.Sprintf("Unknown blocker type: %+v", blockerConfig.BlockerType)
    return ErrorResponse(http.StatusInternalServerError, ErrorTag_Invalid_Value, errString, isAfterTransaction)
  }
  bValid, errorMsg := validator.ValidateWithName(&req, customer, name)
  if !bValid {
    return ErrorResponse(http.StatusBadRequest, ErrorTag_Bad_Attribute, errorMsg, isAfterTransaction)
  }

  return WithTransaction(func (tx *db.Tx) (Response, error) {
    return WithClient(tx, customer, cuid, func (client *data_models.Client) (Response, error) {
      isAfterTransaction = false
      // If the request contains 'cuid' and 'cdid' but DOTS server doesn't maintain 'cdid' for this 'cuid', DOTS server will response 403 Forbidden
      if cdid != "" && (client.Cdid == nil || ( client.Cdid != nil && cdid != *client.Cdid)) {
        errMsg := fmt.Sprintf("Dots server does not maintain a 'cdid' for client with cuid = %+v", client.Cuid)
        log.Error(errMsg)
        return ErrorResponse(http.StatusForbidden, ErrorTag_Access_Denied, errMsg, isAfterTransaction)
      }
      alias := req.Aliases.Alias[0]
      e, err := data_models.FindAliasByName(tx, client, alias.Name, now)
      if err != nil {
        errMsg = fmt.Sprintf("Failed to get alias with name = %+v", alias.Name)
        return ErrorResponse(http.StatusInternalServerError, ErrorTag_Operation_Failed, errMsg, isAfterTransaction)
      }
      alias.TargetPrefix = data_models.RemoveOverlapIPPrefix(alias.TargetPrefix)
      status := http.StatusCreated
      if e == nil {
        t := data_models.NewAlias(client, alias, now, defaultAliasLifetime)
        e = &t
      } else {
        e.Alias = types.Alias(alias)
        e.ValidThrough = now.Add(defaultAliasLifetime)
        status = http.StatusNoContent
      }
      err = e.Save(tx)
      if err != nil {
        errMsg = fmt.Sprintf("Failed to save alias with name = %+v", alias.Name)
        return ErrorResponse(http.StatusInternalServerError, ErrorTag_Operation_Failed, errMsg, isAfterTransaction)
      }

      // Add alias to check expired
      data_models.AddActiveAliasRequest(e.Id, e.Client.Id, e.Alias.Name, e.ValidThrough)

      return EmptyResponse(status)
    })
  })
}

/*
 * Validate aliases: have they been created by the client
 * parameter:
 *  customer request source Customer
 *  cuid client identifier
 *  aliases a list of alias-name
 * return:
 *  res: a list of alias data corresponding to alias-name
 *  err: Error occur in validation progress
 */
func GetDataAliasesByName(customer *models.Customer, cuid string, aliases []string) (res types.Aliases, err error) {
  now := time.Now()
  for _, name := range aliases {
    _, err = WithTransaction(func (tx *db.Tx) (Response, error) {
      return WithClient(tx, customer, cuid, func (client *data_models.Client) (_ Response, err error) {
      alias, err := data_models.FindAliasByName(tx, client, name, now)
      if err != nil {
        return
      }
      if alias == nil {
        log.Warnf("Alias with name: %+v has not been created by client: %+v", name, cuid)
        res.Alias = nil
        return
      }

      ta, err := alias.ToTypesAlias(now)
      if err != nil {
        return
      }

      res.Alias = append(res.Alias, *ta)
      return
      })
    })
  }
  return
}