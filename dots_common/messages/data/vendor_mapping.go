package data_messages

import (
	"fmt"
	"strconv"
	"github.com/nttdots/go-dots/dots_server/models"
	types "github.com/nttdots/go-dots/dots_common/types/data"
)

type VendorMappingRequest struct {
	VendorMapping types.VendorMapping `json:"ietf-dots-mapping:vendor-mapping"`
}

type VendorMappingResponse struct {
	VendorMapping types.VendorMapping `json:"ietf-dots-mapping:vendor-mapping"`
}

// Validate with vendor-id (Put request)
func ValidateWithVendorId(vendorId int, req *VendorMappingRequest) (errMsg string) {
	if len(req.VendorMapping.Vendor) != 1{
		errMsg = fmt.Sprintf("Body Data Error : Have multiple 'vendors' (%d)", len(req.VendorMapping.Vendor))
		return
	}
	vendor := req.VendorMapping.Vendor[0]
	if vendor.VendorId != nil && int(*vendor.VendorId) != vendorId {
		errMsg = fmt.Sprintf("Request/URI vendor-id mismatch : (%v) / (%v)", int(*vendor.VendorId), vendorId)
		return
	}
	return
}

// Validate vendor-mapping (Post/Put request)
func ValidateVendorMapping(req *VendorMappingRequest) (errMsg string) {
	for _, vendor := range req.VendorMapping.Vendor {
		if vendor.VendorId == nil {
			errMsg = fmt.Sprintln("Missing 'vendor-id' required attribute")
			return
		}
		if vendor.DescriptionLang != nil {
			_, errMsg = models.ValidateDescriptionLang(*vendor.DescriptionLang)
			return
		}
		if vendor.LastUpdated == nil {
			errMsg = fmt.Sprintln("Missing 'last-updated' required attribute")
			return
		}
		if _, err := strconv.ParseUint(*vendor.LastUpdated, 10, 64); err != nil {
			errMsg = fmt.Sprintln("The type of 'last-updated' is not uint")
			return
		}
		for _, attack := range vendor.AttackMapping {
			if attack.AttackId == nil {
				errMsg = fmt.Sprintln("Missing 'attack-id' required attribute")
				return
			}
			if attack.AttackDescription == nil {
				errMsg = fmt.Sprintln("Missing 'attack-description' required attribute")
				return
			}
		}
	}
	return
}