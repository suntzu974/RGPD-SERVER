package main

import (
	"time"
)

func init() {}

type Configuration struct {
	Database struct {
		Server   string `json:"server"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
	} `json:"database"`
	Port int    `json:"port"`
	Log  string `json:"log"`
}

type ResponseConsent struct {
	Consent  Consent  `json:"consent"`
	Customer Customer `json:"customer"`
}
type RequestCustomer struct {
	Siret string `json:"siret"`
}
type Stock struct {
	Reference   string  `json:reference`
	Designation string  `json:designation`
	Famille     string  `json:famille`
	Gencod      string  `json:gencod`
	Quantite    float64 `json:quantite`
}
type YTMPBPC struct {
	YLIN_0         string `json:"YLIN_0"`
	BCGCOD_0       string `json:"BCGCOD_0"`
	BPCNUM_0       string `json:"BPCNUM_0"`
	BPCSTA_0       string `json:"BPCSTA_0"`
	BPRNAM_0       string `json:"BPRNAM_0"`
	BPRNAM_1       string `json:"BPRNAM_1"`
	BPRSHO_0       string `json:"BPRSHO_0"`
	BPRLOG_0       string `json:"BPRLOG_0"`
	CRN_0          string `json:"CRN_0"`
	NAF_0          string `json:"NAF_0"`
	CRY_0          string `json:"CRY_0"`
	CUR_0          string `json:"CUR_0"`
	VACBPR_0       string `json:"VACBPR_0"`
	PTE_0          string `json:"PTE_0"`
	ACCCOD_0       string `json:"ACCCOD_0 "`
	TSCCOD_0       string `json:"TSCCOD_0"`
	TSCCOD_1       string `json:"TSCCOD_1"`
	OSTAUZ_0       string `json:"OSTAUZ_0"`
	REP_0          string `json:"REP_0"`
	REP_1          string `json:"REP_1"`
	YBCG_COMPT_0   string `json:"YBCG_COMPT_0"`
	YBPC_RECOUVR_0 string `json:"YBPC_RECOUVR_0"`
	YCATCPT_0      string `json:"YCATCPT_0"`
	YSCATCPT_0     string `json:"YSCATCPT_0"`
	BPAADD_0       string `json:"BPAADD_0"`
	BPADES_0       string `json:"BPADES_0"`
	BPAADDLIG_0    string `json:"BPAADDLIG_0"`
	BPAADDLIG_1    string `json:"BPAADDLIG_1"`
	BPAADDLIG_2    string `json:"BPAADDLIG_2"`
	POSCOD_0       string `json:"POSCOD_0"`
	CTY_0          string `json:"CTY_0"`
	BCRY_0         string `json:"BCRY_0"`
	TEL_0          string `json:"TEL_0"`
	TEL_1          string `json:"TEL_1 "`
	WEB_0          string `json:"WEB_0"`
}
type Customer struct {
	Reference string `json:"Reference"`
	Name      string `json:"Name"`
	Raison    string `json:"-"`
	Sigle     string `json:"-"`
	Identity  string `json:"Identity"`
	Street    string `json:"Street"`
	Address   string `json:"Address"`
	Postcod   string `json:"Postcod"`
	Town      string `json:"Town"`
	Country   string `json:"Country"`
	Phone     string `json:"Phone"`
	Email     string `json:"Email"`
}
type CustomerConsent struct {
	Customer               Customer  `json:"customer"`
	Siret                  string    `json:"Siret"`
	UsingGeneralConditions bool      `json:"UsingGeneralConditions"`
	Newsletters            bool      `json:"Newsletters"`
	CommercialOffersByMail bool      `json:"CommercialOffersByMail"`
	CommercialOffersBySms  bool      `json:"CommercialOffersBySms"`
	CommercialOffersByPost bool      `json:"CommercialOffersByPost"`
	Signature              string    `json:"Signature"`
	CreatedAt              time.Time `json:"CreatedAt"`
}

func (c *CustomerConsent) CreatedCustomer() bool {
	if len(c.Customer.Name) > 0 {
		return true
	} else {
		return false
	}
}

type Consent struct {
	Siret                  string `json:"Siret"`
	UsingGeneralConditions bool   `json:"UsingGeneralConditions"`
	Newsletters            bool   `json:"Newsletters"`
	CommercialOffersByMail bool   `json:"CommercialOffersByMail"`
	CommercialOffersBySms  bool   `json:"CommercialOffersBySms"`
	CommercialOffersByPost bool   `json:"CommercialOffersByPost"`
	Signature              string `json:"Signature"`
	//CreatedAt              time.Time `json:"CreatedAt"`
}
