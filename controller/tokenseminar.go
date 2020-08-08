package controller

import (
	"database/sql"
	"zoomgateway/localtools"

	"github.com/huandu/go-sqlbuilder"

	"github.com/valyala/fasthttp"
)

func TokenSeminarController(dbConn *sql.DB, ctx *fasthttp.RequestCtx, pagesettings map[string]interface{}, zoomsettings map[string]interface{}) {
	idpeserta := ctx.UserValue("nimmhs").(string)
	namapeserta := ctx.UserValue("namamhs").(string)
	idsesi := ctx.UserValue("idsesi").(string)

	sqlb := sqlbuilder.NewSelectBuilder()
	sqlb.Select("akun.apikey", "akun.apisecret", "sesi.meetingid", "sesi.password")
	sqlb.From("pesertapersesi")
	sqlb.Join("sesi", "sesi.id = pesertapersesi.sesi_id")
	sqlb.Join("akun", "sesi.akun_id = akun.id")
	sqlb.Where(sqlb.And(
		sqlb.E("pesertapersesi.peserta_id", idpeserta),
		sqlb.E("pesertapersesi.sesi_id", idsesi),
	))

	var apikey, apisecret, meetid, meetpass string
	a, b := sqlb.Build()
	err := dbConn.QueryRow(a, b...).Scan(&apikey, &apisecret, &meetid, &meetpass)
	if err != nil {
		localtools.LogThisError(ctx, err.Error())
		return
	}

	type payloadreply struct {
		Signature     string `json:"signature"`
		MeetingNumber string `json:"meetingNumber"`
		UserName      string `json:"userName"`
		APIKey        string `json:"apiKey"`
		APISecret     string `json:"hash"`
		Password      string `json:"password"`
	}

	var reply payloadreply
	reply.Signature = "akjsUHakjSAUh"
	reply.APIKey = apikey
	reply.MeetingNumber = meetid
	reply.Password = meetpass
	reply.UserName = namapeserta
	reply.APISecret = apisecret

	localtools.DoJSONWrite(ctx, 200, reply)

}
