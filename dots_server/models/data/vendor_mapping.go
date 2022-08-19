package data_models

import (
	"strconv"
	"github.com/nttdots/go-dots/dots_server/db"
	"github.com/nttdots/go-dots/dots_server/db_models/data"
	log "github.com/sirupsen/logrus"
	types "github.com/nttdots/go-dots/dots_common/types/data"
)

type Vendor struct {
	Id              int64
	ClientId        int64
	VendorId        int
	VendorName      string
	DescriptionLang string
	LastUpdated     uint64
	AttackMapping   []AttackMapping
}

type AttackMapping struct {
	Id                int64
	AttackId          int
	AttackDescription string
}

type Vendors []Vendor

// New vendor-mapping
func NewVendorMapping(clientId int64, bodyData types.Vendor) Vendor {
	vendor := Vendor{}
	lastUpdated, _ := strconv.ParseUint(*bodyData.LastUpdated, 10, 64)
	vendor.ClientId        = clientId
	vendor.VendorId        = int(*bodyData.VendorId)
	vendor.VendorName      = *bodyData.VendorName
	vendor.DescriptionLang = *bodyData.DescriptionLang
	vendor.LastUpdated     = lastUpdated
	for _, v := range bodyData.AttackMapping {
		attackMapping := AttackMapping{}
		attackMapping.AttackId          = int(*v.AttackId)
		attackMapping.AttackDescription = *v.AttackDescription
		vendor.AttackMapping = append(vendor.AttackMapping, attackMapping)
	}
	return vendor
}

// Insert vendor-mapping into DB
func (vendor *Vendor) Save(tx *db.Tx) error {
	v := data_db_models.VendorMapping{}
	v.Id              = vendor.Id
	v.DataClientId    = vendor.ClientId
	v.VendorId        = vendor.VendorId
	v.VendorName      = vendor.VendorName
	v.DescriptionLang = vendor.DescriptionLang
	v.LastUpdated     = vendor.LastUpdated

	if v.Id == 0 {
		// Register vendor-mapping
		_, err := tx.Session.Insert(&v)
		if err != nil {
			log.WithError(err).Error("Insert() vendor-mapping failed.")
			return err
		}
	} else {
		// Update vendor-mapping
		_, err := tx.Session.ID(v.Id).Update(&v)
		if err != nil {
			log.WithError(err).Errorf("Update() vendor-mapping failed.")
			return err
		}
		// Delete attack-mapping
		_, err =tx.Session.Delete(&data_db_models.AttackMapping{VendorMappingId: vendor.Id})
		if err != nil {
			log.WithError(err).Errorf("Delete() attack-mapping failed.")
			return err
		}
	}
	// Register attack-mapping
	for _, attack := range vendor.AttackMapping {
		a := data_db_models.AttackMapping{}
		a.VendorMappingId   = v.Id
		a.AttackId          = attack.AttackId
		a.AttackDescription = attack.AttackDescription
		_, err := tx.Session.Insert(&a)
		if err != nil {
			log.WithError(err).Error("Insert() attack-mapping failed.")
			return err
		}
	}
	return nil
}

// Find vendor-mapping by vendor-id
func FindVendorByVendorId(tx *db.Tx, clientId int64, vendorId int) (Vendor, error) {
	vendor := Vendor{}
	dbVendor := data_db_models.VendorMapping{}
	_, err := tx.Session.Where("data_client_id = ? AND vendor_id = ?", clientId, vendorId).Get(&dbVendor)
	if err != nil {
		return vendor, err
	}
	vendor.Id              = dbVendor.Id
	vendor.ClientId        = dbVendor.DataClientId
	vendor.VendorId        = dbVendor.VendorId
	vendor.VendorName      = dbVendor.VendorName
	vendor.DescriptionLang = dbVendor.DescriptionLang
	vendor.LastUpdated     = dbVendor.LastUpdated
	// Get attack-mapping
	dbAttackList := []data_db_models.AttackMapping{}
	err = tx.Session.Where("vendor_mapping_id = ?", dbVendor.Id).Find(&dbAttackList)
	if err != nil {
		log.WithError(err).Error("Get() attack-mapping failed.")
		return vendor, err
	}
	for _, dbAttack := range dbAttackList {
		attack := AttackMapping{
			Id:                dbAttack.Id,
			AttackId:          dbAttack.AttackId,
			AttackDescription: dbAttack.AttackDescription,
		}
		vendor.AttackMapping = append(vendor.AttackMapping, attack)
	}
	return vendor, nil
}

// Find vendor-mapping by client-id
func FindVendorMappingByClientId(tx *db.Tx, clientId *int64) (Vendors, error) {
	vendorList := make(Vendors, 0)
	// Get vendor-mapping
	dbVendorList := []data_db_models.VendorMapping{}
	if clientId != nil {
		err := tx.Session.Where("data_client_id = ?", clientId).Find(&dbVendorList)
		if err != nil {
			log.WithError(err).Error("Get() vendor-mapping failed.")
			return nil, err
		}
	} else {
		err := tx.Session.Where("data_client_id = ?", 0).Find(&dbVendorList)
		if err != nil {
			log.WithError(err).Error("Get() vendor-mapping failed.")
			return nil, err
		}
	}
	for _, dbVendor := range dbVendorList {
		vendor := Vendor{}
		vendor.ClientId        = dbVendor.DataClientId
		vendor.VendorId        = dbVendor.VendorId
		vendor.VendorName      = dbVendor.VendorName
		vendor.DescriptionLang = dbVendor.DescriptionLang
	    vendor.LastUpdated     = dbVendor.LastUpdated
		// Get attack-mapping
		dbAttackList := []data_db_models.AttackMapping{}
		err := tx.Session.Where("vendor_mapping_id = ?", dbVendor.Id).Find(&dbAttackList)
		if err != nil {
			log.WithError(err).Error("Get() attack-mapping failed.")
			return nil, err
		}
		for _, dbAttack := range dbAttackList {
			attack := AttackMapping{
				Id:                dbAttack.Id,
				AttackId:          dbAttack.AttackId,
				AttackDescription: dbAttack.AttackDescription,
			}
			vendor.AttackMapping = append(vendor.AttackMapping, attack)
		}
		vendorList = append(vendorList, vendor)
	}
	return vendorList, nil
}

// Convert vendor-mapping to types vendor-mapping
func (vendors Vendors) ToTypesVendorMapping(depth *int) (*types.VendorMapping) {
	vendorList := types.VendorMapping{}
	if depth == nil || *depth > 2 {
		for _, v := range vendors {
			vendor := types.Vendor{}
			vendorId := uint32(v.VendorId)
			lastUpdated := strconv.FormatUint(v.LastUpdated, 10)
			vendor.VendorId        = &vendorId
			vendor.VendorName      = &v.VendorName
			vendor.DescriptionLang = &v.DescriptionLang
			vendor.LastUpdated     = &lastUpdated
			if depth == nil || *depth > 3 {
				for _, a := range v.AttackMapping {
					attack := types.AttackMapping{}
					attackId := uint32(a.AttackId)
					attackDescription := a.AttackDescription
					attack.AttackId   = &attackId
					attack.AttackDescription = &attackDescription
					vendor.AttackMapping = append(vendor.AttackMapping, attack)
				}
			} else {
				vendor.AttackMapping = []types.AttackMapping{}
			}
			vendorList.Vendor = append(vendorList.Vendor, vendor)
		}
	} else {
		vendorList.Vendor = []types.Vendor{}
	}
	return &vendorList
}