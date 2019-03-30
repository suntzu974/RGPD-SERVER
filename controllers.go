package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

func GetDatabase(c *Configuration) (db *sql.DB) {

	var server = c.Database.Server
	var port = c.Database.Port
	var user = c.Database.User
	var password = c.Database.Password
	//	var database = c.Database.Database

	var err error
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d",
		server, user, password, port)
	db, err = sql.Open("sqlserver", connString)
	if err != nil {
		log.Printf("Error creating connection pool: %s", err.Error())
	}
	return db
}

func SelectVersion(db *sql.DB) string {
	ctx := context.Background()
	err := db.PingContext(ctx)
	if err != nil {
		log.Printf("Error pinging database: %s", err.Error())
	}
	var result string
	err = db.QueryRowContext(ctx, "SELECT @@version").Scan(&result)
	if err != nil {
		log.Printf("Scan failed: %s", err.Error())
	}
	return result + "\n"
}
func (c *Configuration) homeHandler(w http.ResponseWriter, r *http.Request) {
	var rgpdTemplate *template.Template
	var err error
	switch os := runtime.GOOS; os {
	case "darwin":
		log.Printf("Platform from html  %s.\n", os)
		rgpdTemplate, err = template.ParseFiles("/Users/suntzu974/GoProjects/RGPD-SERVER/rgpd.html")
	case "linux":
		log.Printf("Platform from html  %s.\n", os)
		rgpdTemplate, err = template.ParseFiles("/home/jeannick/PROJECTS/RGPD-SERVER/rgpd.html")
	default:
		log.Printf("Platform from html  %s.\n", os)
		rgpdTemplate, err = template.ParseFiles("rgpd.html")
	}

	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "text/html")

	consents, _, err := AllConsents(GetDatabase(c))
	err = rgpdTemplate.Execute(w, consents)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func ReadStockFromSofarem(db *sql.DB) ([]Stock, []Stock, int, error) {
	var count int = 0
	stocks := []Stock{}
	stocks_hometech := []Stock{}
	ctx := context.Background()
	err := db.PingContext(ctx)
	if err != nil {
		log.Printf("Error pinging database: %s", err.Error())
	}

	request := `SELECT ITP.ITMREFBPS_0,ITM.ITMDES1_0,YFA2.YFA_LIB_0, ITM.EANCOD_0,
	SUM(STO.QTYSTU_0 - ITV.SALSTO_0)
	FROM  [x3prod].[RAVPROD].[ITMFACILIT] AS ITF
	LEFT JOIN [x3prod].[RAVPROD].[ITMMASTER] AS ITM ON ITM.ITMREF_0 = ITF.ITMREF_0
	LEFT JOIN [x3prod].[RAVPROD].[ITMBPS] AS ITP ON ITP.ITMREF_0 = ITF.ITMREF_0 AND ITP.PIO_0 = 0
	JOIN [x3prod].[RAVPROD].[STOCK] AS STO ON ITF.ITMREF_0 = STO.ITMREF_0 AND STO.STOFCY_0 = 076
	AND LOC_0 in ('A2A','B1A') AND STA_0 <> 'R'
	LEFT JOIN   [x3prod].[RAVPROD].[ITMMVT] AS ITV ON ITV.ITMREF_0 = ITF.ITMREF_0 AND ITV.STOFCY_0 in (076)
	LEFT JOIN   [x3prod].[RAVPROD].[BPSUPPLIER] AS BPS ON BPS.BPSNUM_0 = ITP.BPSNUM_0
	LEFT JOIN   [x3prod].[RAVPROD].[YFAMART] AS YFA0 ON ITM.YITM_FAM2_0 = YFA0.YFA_CODE_0 AND YFA0.YNF_NIV_0 = 1 AND YFA0.YNF_REG_0 = 2
	LEFT JOIN   [x3prod].[RAVPROD].[YFAMART] AS YFA1 ON ITM.YITM_FAM2_1 = YFA1.YFA_CODE_0 AND YFA1.YNF_NIV_0 = 2 AND YFA1.YNF_REG_0 = 2
	LEFT JOIN   [x3prod].[RAVPROD].[YFAMART] AS YFA2 ON ITM.YITM_FAM2_2 = YFA2.YFA_CODE_0 AND YFA2.YNF_NIV_0 = 3 AND YFA2.YNF_REG_0 = 2
	LEFT JOIN   [x3prod].[RAVPROD].[YFAMART] AS YFA3 ON ITM.YITM_FAM2_3 = YFA3.YFA_CODE_0 AND YFA3.YNF_NIV_0 = 4 AND YFA3.YNF_REG_0 = 2
	WHERE ITF.STOFCY_0 in (076) AND (STO.QTYSTU_0 - ITV.SALSTO_0) <> 0
	GROUP BY ITP.ITMREFBPS_0,ITM.YITM_ANRAV_0,ITM.ITMDES1_0,YFA2.YFA_LIB_0, ITM.EANCOD_0
`
	tsql := fmt.Sprintf("%s", request)
	rows, err := db.QueryContext(ctx, tsql)
	if err != nil {
		log.Printf("Error reading rows: %s", err.Error())
		return stocks, stocks_hometech, -1, err
	}

	defer rows.Close()
	for rows.Next() {
		stock := Stock{}
		// Get values from row.
		err := rows.Scan(&stock.Reference, &stock.Designation,
			&stock.Famille, &stock.Gencod, &stock.Quantite)
		if err != nil {
			log.Printf("Error reading Stock : %s", err.Error())
			return stocks, stocks_hometech, -1, err
		}
		stocks = append(stocks, stock)
		count++
	}
	// Hometech
	request0 := `SELECT ITP.ITMREFBPS_0,ITM.ITMDES1_0,YFA2.YFA_LIB_0, ITM.EANCOD_0,
	SUM(STO.QTYSTU_0 - ITV.SALSTO_0)
	FROM  [x3prod].[RAVPROD].[ITMFACILIT] AS ITF
	LEFT JOIN [x3prod].[RAVPROD].[ITMMASTER] AS ITM ON ITM.ITMREF_0 = ITF.ITMREF_0
	LEFT JOIN [x3prod].[RAVPROD].[ITMBPS] AS ITP ON ITP.ITMREF_0 = ITF.ITMREF_0 AND ITP.PIO_0 = 0
	JOIN [x3prod].[RAVPROD].[STOCK] AS STO ON ITF.ITMREF_0 = STO.ITMREF_0 AND STO.STOFCY_0 = 041
	AND LOC_0 in ('A2A','B1A') AND STA_0 <> 'R'
	LEFT JOIN   [x3prod].[RAVPROD].[ITMMVT] AS ITV ON ITV.ITMREF_0 = ITF.ITMREF_0 AND ITV.STOFCY_0 in (041)
	LEFT JOIN   [x3prod].[RAVPROD].[BPSUPPLIER] AS BPS ON BPS.BPSNUM_0 = ITP.BPSNUM_0
	LEFT JOIN   [x3prod].[RAVPROD].[YFAMART] AS YFA0 ON ITM.YITM_FAM2_0 = YFA0.YFA_CODE_0 AND YFA0.YNF_NIV_0 = 1 AND YFA0.YNF_REG_0 = 2
	LEFT JOIN   [x3prod].[RAVPROD].[YFAMART] AS YFA1 ON ITM.YITM_FAM2_1 = YFA1.YFA_CODE_0 AND YFA1.YNF_NIV_0 = 2 AND YFA1.YNF_REG_0 = 2
	LEFT JOIN   [x3prod].[RAVPROD].[YFAMART] AS YFA2 ON ITM.YITM_FAM2_2 = YFA2.YFA_CODE_0 AND YFA2.YNF_NIV_0 = 3 AND YFA2.YNF_REG_0 = 2
	LEFT JOIN   [x3prod].[RAVPROD].[YFAMART] AS YFA3 ON ITM.YITM_FAM2_3 = YFA3.YFA_CODE_0 AND YFA3.YNF_NIV_0 = 4 AND YFA3.YNF_REG_0 = 2
	WHERE ITF.STOFCY_0 in (041) AND (STO.QTYSTU_0 - ITV.SALSTO_0) <> 0
	GROUP BY ITP.ITMREFBPS_0,ITM.YITM_ANRAV_0,ITM.ITMDES1_0,YFA2.YFA_LIB_0, ITM.EANCOD_0
`
	tsql = fmt.Sprintf("%s", request0)

	rows, err = db.QueryContext(ctx, tsql)
	if err != nil {
		log.Printf("Error reading rows: %s", err.Error())
		return stocks, stocks_hometech, -1, err
	}

	defer rows.Close()
	for rows.Next() {
		stock := Stock{}
		// Get values from row.
		err := rows.Scan(&stock.Reference, &stock.Designation,
			&stock.Famille, &stock.Gencod, &stock.Quantite)
		if err != nil {
			log.Printf("Error reading Stock : %s", err.Error())
			return stocks, stocks_hometech, -1, err
		}
		stocks_hometech = append(stocks_hometech, stock)
		count++
	}

	log.Printf("Rows read for Stocks  %d.\n", count)
	return stocks, stocks_hometech, count, nil

}
func ReadCustomer(db *sql.DB, query string) (Customer, int, error) {
	var count int = 0
	customer := Customer{}
	ctx := context.Background()

	// Check if database is alive.
	err := db.PingContext(ctx)
	if err != nil {
		log.Printf("Error pinging database: %s", err.Error())
	}

	request := `SELECT [BPCNUM_0],[BPCNAM_0],[CRN_0]
      ,[BPAADDLIG_0],[BPAADDLIG_1],[POSCOD_0],[CTY_0],[BPARTNER].[CRY_0]
         ,[TEL_0],[WEB_0]
  		 	FROM [x3prod].[RAVPROD].[BPCUSTOMER],[x3prod].[RAVPROD].[BPADDRESS],[x3prod].[RAVPROD].[BPARTNER]
  	 		WHERE [x3prod].[RAVPROD].[BPADDRESS].[BPANUM_0] = [x3prod].[RAVPROD].[BPCUSTOMER].[BPCNUM_0]
       	AND [x3prod].[RAVPROD].[BPARTNER].[CRN_0]   =   @query
       	AND [x3prod].[RAVPROD].[BPCUSTOMER].[BPAINV_0] = [x3prod].[RAVPROD].[BPADDRESS].[BPAADD_0]
       	AND [x3prod].[RAVPROD].[BPADDRESS].[BPATYP_0] = 1
				AND [x3prod].[RAVPROD].[BPCUSTOMER].[BPCNUM_0] = [x3prod].[RAVPROD].[BPARTNER].[BPRNUM_0]
       	ORDER BY [x3prod].[RAVPROD].[BPCUSTOMER].[BPCNAM_0] ASC;`
	tsql := fmt.Sprintf("%s", request)
	log.Printf("Request Customer %s", tsql)

	// Execute query
	rows, err := db.QueryContext(ctx, tsql, sql.Named("query", query))
	if err != nil {
		log.Printf("Error reading rows: %s", err.Error())
		return customer, -1, err
	}

	defer rows.Close()

	// Iterate through the result set.
	for rows.Next() {

		// Get values from row.
		err := rows.Scan(&customer.Reference, &customer.Name, &customer.Identity,
			&customer.Street, &customer.Address, &customer.Postcod, &customer.Town,
			&customer.Country, &customer.Phone, &customer.Email)
		if err != nil {
			log.Printf("Error reading rows: %s", err.Error())
			return customer, -1, err
		}
		count++
	}
	return customer, count, nil
}
func CreateCustomer(db *sql.DB, c Customer) (int64, error) {
	bpc := YTMPBPC{}
	bpc.YLIN_0 = "1"
	bpc.BCGCOD_0 = "92"
	bpc.BPCNUM_0 = ""
	bpc.BPCSTA_0 = "2"
	bpc.BPRNAM_0 = c.Name
	bpc.BPRNAM_1 = c.Raison
	bpc.BPRSHO_0 = "CCOMPTOIR"
	bpc.BPRLOG_0 = c.Sigle
	bpc.CRN_0 = c.Identity
	bpc.NAF_0 = ""
	bpc.CRY_0 = c.Country
	bpc.CUR_0 = "EUR"
	bpc.VACBPR_0 = "DOM"
	bpc.PTE_0 = "CHQCPT"
	bpc.ACCCOD_0 = "CPT"
	bpc.TSCCOD_0 = ""
	bpc.TSCCOD_1 = ""
	bpc.OSTAUZ_0 = "1"
	bpc.REP_0 = "KWONGCHEONG"
	bpc.REP_1 = ""
	bpc.YBCG_COMPT_0 = "2"
	bpc.YBPC_RECOUVR_0 = "50"
	bpc.YCATCPT_0 = "1"
	bpc.YSCATCPT_0 = "1"
	bpc.BPAADD_0 = "A1"
	bpc.BPADES_0 = ""
	bpc.BPAADDLIG_0 = c.Street
	bpc.BPAADDLIG_1 = c.Address
	bpc.BPAADDLIG_2 = ""
	bpc.POSCOD_0 = c.Postcod
	bpc.CTY_0 = c.Town
	bpc.BCRY_0 = c.Country
	bpc.TEL_0 = c.Phone
	bpc.TEL_1 = ""
	bpc.WEB_0 = c.Email

	request := `INSERT INTO [x3prod].[RAVPROD].[ZET_TMPBPC]
           ([YLIN_0] ,[BCGCOD_0] ,[BPCNUM_0] ,[BPCSTA_0] ,[BPRNAM_0]
           ,[BPRNAM_1] ,[BPRSHO_0] ,[BPRLOG_0] ,[CRN_0] ,[NAF_0]
           ,[CRY_0] ,[CUR_0] ,[VACBPR_0]
           ,[PTE_0] ,[ACCCOD_0],[TSCCOD_0],[TSCCOD_1]
					 ,[OSTAUZ_0],[REP_0],[REP_1]
           ,[YBCG_COMPT_0] ,[YBPC_RECOUVR_0],[YCATCPT_0]
           ,[YSCATCPT_0] ,[BPAADD_0] ,[BPADES_0],[BPAADDLIG_0]
           ,[BPAADDLIG_1],[BPAADDLIG_2] ,[POSCOD_0],[CTY_0]
           ,[BCRY_0] ,[TEL_0] ,[TEL_1] , [WEB_0])
     VALUES
           (@YLIN_0 ,@BCGCOD_0 ,@BPCNUM_0 ,@BPCSTA_0 ,@BPRNAM_0
           ,@BPRNAM_1,@BPRSHO_0,@BPRLOG_0 ,@CRN_0 ,@NAF_0
           ,@CRY_0,@CUR_0,@VACBPR_0,@PTE_0,@ACCCOD_0,@TSCCOD_0
           ,@TSCCOD_1,@OSTAUZ_0,@REP_0,@REP_1
           ,@YBCG_COMPT_0 ,@YBPC_RECOUVR_0 ,@YCATCPT_0
           ,@YSCATCPT_0 ,@BPAADD_0 ,@BPADES_0 ,@BPAADDLIG_0 ,@BPAADDLIG_1
           ,@BPAADDLIG_2,@POSCOD_0,@CTY_0,@BCRY_0,@TEL_0 ,@TEL_1
					 ,@WEB_0 )`

	stmt, err := db.Prepare(request)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	ctx := context.Background()
	result, err := stmt.ExecContext(
		ctx,
		sql.Named("YLIN_0", bpc.YLIN_0),
		sql.Named("BCGCOD_0", bpc.BCGCOD_0),
		sql.Named("BPCNUM_0", bpc.BPCNUM_0),
		sql.Named("BPCSTA_0", bpc.BPCSTA_0),
		sql.Named("BPRNAM_0", bpc.BPRNAM_0),
		sql.Named("BPRNAM_1", bpc.BPRNAM_1),
		sql.Named("BPRSHO_0", bpc.BPRSHO_0),
		sql.Named("BPRLOG_0", bpc.BPRLOG_0),
		sql.Named("CRN_0", bpc.CRN_0),
		sql.Named("NAF_0", bpc.NAF_0),
		sql.Named("CRY_0", bpc.CRY_0),
		sql.Named("CUR_0", bpc.CUR_0),
		sql.Named("VACBPR_0", bpc.VACBPR_0),
		sql.Named("PTE_0", bpc.PTE_0),
		sql.Named("ACCCOD_0", bpc.ACCCOD_0),
		sql.Named("TSCCOD_0", bpc.TSCCOD_0),
		sql.Named("TSCCOD_1", bpc.TSCCOD_1),
		sql.Named("OSTAUZ_0", bpc.OSTAUZ_0),
		sql.Named("REP_0", bpc.REP_0),
		sql.Named("REP_1", bpc.REP_1),
		sql.Named("YBCG_COMPT_0", bpc.YBCG_COMPT_0),
		sql.Named("YBPC_RECOUVR_0", bpc.YBPC_RECOUVR_0),
		sql.Named("YCATCPT_0", bpc.YCATCPT_0),
		sql.Named("YSCATCPT_0", bpc.YSCATCPT_0),
		sql.Named("BPAADD_0", bpc.BPAADD_0),
		sql.Named("BPADES_0", bpc.BPADES_0),
		sql.Named("BPAADDLIG_0", bpc.BPAADDLIG_0),
		sql.Named("BPAADDLIG_1", bpc.BPAADDLIG_1),
		sql.Named("BPAADDLIG_2", bpc.BPAADDLIG_2),
		sql.Named("POSCOD_0", bpc.POSCOD_0),
		sql.Named("CTY_0", bpc.CTY_0),
		sql.Named("BCRY_0", bpc.BCRY_0),
		sql.Named("TEL_0", bpc.TEL_0),
		sql.Named("TEL_1", bpc.TEL_1),
		sql.Named("WEB_0", bpc.WEB_0))

	rows, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error in write data %s", err.Error())
		return -1, err
	}
	log.Printf("Rows inserted into YTMPBPC  %d.\n", rows)
	return rows, nil
}
func UpdateConsent(db *sql.DB, consent Consent) (int64, error) {
	request := `UPDATE [x3prod].[RAVPROD].[ZET_CDVRGPD]
   	SET [CONSENT_0] = @conditions
      ,[CONSENT_1] = @newsletters
      ,[CONSENT_2] = @mail
      ,[CONSENT_3] = @sms
      ,[CONSENT_4] = @post
      ,[SIGNCHAR_0] = @signature
      ,[DATE_0] = GETDATE()
 		WHERE [CRN_0] = @siret`
	stmt, err := db.Prepare(request)
	if err != nil {
		log.Printf("Error in write data %s", err.Error())
		return -1, err
	}
	defer stmt.Close()

	ctx := context.Background()
	result, err := stmt.ExecContext(
		ctx,
		sql.Named("siret", consent.Siret),
		sql.Named("conditions", consent.UsingGeneralConditions),
		sql.Named("newsletters", consent.Newsletters),
		sql.Named("mail", consent.CommercialOffersByMail),
		sql.Named("sms", consent.CommercialOffersBySms),
		sql.Named("post", consent.CommercialOffersByPost),
		sql.Named("signature", consent.Signature))

	rows, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error in write data %s", err.Error())
		return -1, err
	}

	log.Printf("Rows updated for Consent %d.\n", rows)
	return rows, nil
}
func CreateConsent(db *sql.DB, consent Consent) (int64, error) {
	request := `INSERT INTO [x3prod].[RAVPROD].[ZET_CDVRGPD]
           ([CRN_0]
           ,[CONSENT_0]
           ,[CONSENT_1]
           ,[CONSENT_2]
           ,[CONSENT_3]
           ,[CONSENT_4]
           ,[SIGNCHAR_0]
           ,[DATE_0])
     VALUES
           (@siret,@conditions,@newsletters,@mail,@sms,@post,@signature,@created)`
	stmt, err := db.Prepare(request)
	if err != nil {
		log.Printf("Error in write data %s", err.Error())
		return -1, err
	}
	defer stmt.Close()

	ctx := context.Background()
	result, err := stmt.ExecContext(
		ctx,
		sql.Named("siret", consent.Siret),
		sql.Named("conditions", consent.UsingGeneralConditions),
		sql.Named("newsletters", consent.Newsletters),
		sql.Named("mail", consent.CommercialOffersByMail),
		sql.Named("sms", consent.CommercialOffersBySms),
		sql.Named("post", consent.CommercialOffersByPost),
		sql.Named("signature", consent.Signature),
		sql.Named("created", time.Now()))

	rows, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error in write data %s", err.Error())
		return -1, err
	}

	log.Printf("Rows created for Consent %d.\n", rows)
	return rows, nil
}
func AllConsents(db *sql.DB) ([]CustomerConsent, int, error) {
	var count int = 0
	consents := []CustomerConsent{}
	ctx := context.Background()

	// Check if database is alive.
	err := db.PingContext(ctx)
	if err != nil {
		log.Printf("Error pinging database: %s", err.Error())
	}

	request := `SELECT [CRN_0]
			,[CONSENT_0],[CONSENT_1],[CONSENT_2],[CONSENT_3],[CONSENT_4]
				 ,[SIGNCHAR_0] ,[DATE_0]
				FROM [x3prod].[RAVPROD].[ZET_CDVRGPD]
				ORDER BY [x3prod].[RAVPROD].[ZET_CDVRGPD].[CRN_0] ASC;`
	tsql := fmt.Sprintf("%s", request)

	// Execute query
	rows, err := db.QueryContext(ctx, tsql)
	if err != nil {
		log.Printf("Error reading rows: %s ", err.Error())
		return consents, -1, err
	}

	defer rows.Close()

	// Iterate through the result set.
	for rows.Next() {
		customerConsent := CustomerConsent{}
		// Get values from row.
		err := rows.Scan(&customerConsent.Siret, &customerConsent.UsingGeneralConditions,
			&customerConsent.Newsletters,
			&customerConsent.CommercialOffersByMail, &customerConsent.CommercialOffersBySms,
			&customerConsent.CommercialOffersByPost, &customerConsent.Signature, &customerConsent.CreatedAt)
		if err != nil {
			log.Printf("Error reading Customer consent : %s", err.Error())
			return consents, -1, err
		}
		// Read Customer
		customer, _, err := ReadCustomer(db, customerConsent.Siret)
		if err != nil {
			log.Printf("Error Reading customer: %s", err.Error())
		}
		customerConsent.Customer = customer
		consents = append(consents, customerConsent)
		count++
	}
	log.Printf("Rows read for Consents  %d.\n", count)
	return consents, count, nil

}
func ReadConsent(db *sql.DB, query string) (ResponseConsent, int, error) {
	var count int = 0
	consent := Consent{}
	responseConsent := ResponseConsent{}

	ctx := context.Background()

	// Check if database is alive.
	err := db.PingContext(ctx)
	if err != nil {
		log.Printf("Error in write data %s", err.Error())
	}

	request := `SELECT [CRN_0]
      ,[CONSENT_0],[CONSENT_1],[CONSENT_2],[CONSENT_3],[CONSENT_4]
         ,[SIGNCHAR_0] /*,[DATE_0]*/
  		 	FROM [x3prod].[RAVPROD].[ZET_CDVRGPD]
  	 		WHERE [x3prod].[RAVPROD].[ZET_CDVRGPD].[CRN_0]  =  @query
       	ORDER BY [x3prod].[RAVPROD].[ZET_CDVRGPD].[CRN_0] ASC;`
	tsql := fmt.Sprintf("%s", request)

	// Execute query
	rows, err := db.QueryContext(ctx, tsql, sql.Named("query", query))
	if err != nil {
		log.Printf("Error reading Consent: " + err.Error())
		return responseConsent, -1, err
	}

	defer rows.Close()

	// Iterate through the result set.
	for rows.Next() {

		// Get values from row.
		err := rows.Scan(&consent.Siret, &consent.UsingGeneralConditions,
			&consent.Newsletters,
			&consent.CommercialOffersByMail, &consent.CommercialOffersBySms,
			&consent.CommercialOffersByPost, &consent.Signature /*, &customer.CreatedAt*/)
		if err != nil {
			log.Printf("Error reading Consent %s", err.Error())
			return responseConsent, -1, err
		}
		count++
	}
	if count > 0 {
		// Read Customer
		customer, _, err := ReadCustomer(db, query)
		if err != nil {
			log.Printf("Error Reading Consent: %s", err.Error())
		}
		responseConsent.Customer = customer
	} else {
		// ReadCustomer
		customer, _, err := ReadCustomer(db, query)
		if err != nil {
			log.Printf("Error Reading Consent: %s", err.Error())
		}
		responseConsent.Customer = customer
	}
	responseConsent.Consent = consent
	log.Printf("Rows read for Consent %d.\n", count)
	return responseConsent, count, nil
}

func LoadConfiguration(file string) Configuration {
	var config Configuration
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		log.Printf("%s", err.Error())
		config.Database.Server = ""
		config.Database.Port = 1433
		config.Database.User = ""
		config.Database.Password = ""
		config.Database.Database = ""
		config.Port = 50001
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	log.Printf("Configuration loaded succesfully at Port %d.\n", config.Port)
	return config
}
