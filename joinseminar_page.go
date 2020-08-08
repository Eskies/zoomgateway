package main

import (
	"database/sql"
	"strings"
	"time"
	"zoomgateway/localtools"

	"github.com/huandu/go-sqlbuilder"

	"github.com/valyala/fasthttp"
)

func JoinSeminarController(dbConn *sql.DB, ctx *fasthttp.RequestCtx, pagesettings map[string]interface{}, zoomsettings map[string]interface{}) {
	idpeserta := ctx.UserValue("nimmhs").(string)
	idsesi := ctx.UserValue("idsesi").(string)

	sqlb := sqlbuilder.NewUpdateBuilder()
	sqlb.Update("pesertapersesi")
	sqlb.Set(
		sqlb.Assign("waktulogin", time.Now().Format("2006-01-02 15:04:05")),
		sqlb.Assign("loginip", ctx.RemoteIP().String()),
	)
	sqlb.Where(
		sqlb.E("peserta_id", idpeserta),
		sqlb.E("sesi_id", idsesi),
	)

	a, b := sqlb.Build()
	_, err := dbConn.Exec(a, b...)
	if err != nil {
		localtools.LogThisError(ctx, err.Error())
		return
	}

	output := JoinSeminarTemplate()

	output = strings.ReplaceAll(output, "$page_title", pagesettings["title"].(string))
	output = strings.ReplaceAll(output, "$page_description", pagesettings["description"].(string))
	output = strings.ReplaceAll(output, "$nim", ctx.UserValue("nimmhs").(string))
	output = strings.ReplaceAll(output, "$nama", ctx.UserValue("namamhs").(string))
	output = strings.ReplaceAll(output, "$idsesi", ctx.UserValue("idsesi").(string))

	ctx.WriteString(output)

	ctx.Response.Header.SetContentType("text/html; charset=UTF-8")
	ctx.SetConnectionClose()
	ctx.SetStatusCode(200)

	return
}

func JoinSeminarTemplate() string {
	var bufftemplate string
	bufftemplate = `
	<!DOCTYPE html>
	<html lang="en">

	<head>
		
		<title>$page_title</title>
		<meta name="description" content="$page_description">
		<meta name="author" content="Ngurah Ady Kusuma (github.com/Eskies) (email: ady_kusuma@stikom-bali.ac.id">
		
		<meta charset="utf-8" />
		<link type="text/css" rel="stylesheet" href="https://source.zoom.us/1.7.10/css/bootstrap.css" />
		<link type="text/css" rel="stylesheet" href="https://source.zoom.us/1.7.10/css/react-select.css" />
		<meta name="format-detection" content="telephone=no">
		<meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
	</head>
	<body>
		<!-- import ZoomMtg dependencies -->
		<script src="https://source.zoom.us/1.7.10/lib/vendor/react.min.js"></script>
		<script src="https://source.zoom.us/1.7.10/lib/vendor/react-dom.min.js"></script>
		<script src="https://source.zoom.us/1.7.10/lib/vendor/redux.min.js"></script>
		<script src="https://source.zoom.us/1.7.10/lib/vendor/redux-thunk.min.js"></script>
		<script src="https://source.zoom.us/1.7.10/lib/vendor/jquery.min.js"></script>
		<script src="https://source.zoom.us/1.7.10/lib/vendor/lodash.min.js"></script>
		<script src="https://source.zoom.us/zoom-meeting-1.7.10.min.js"></script>

		
		<script>
			$(window).on('load', function() {
				console.log(JSON.stringify(ZoomMtg.checkSystemRequirements()));
				$.get("/tokenseminar/$idsesi")
				.done(function(data){
					ZoomMtg.setZoomJSLib('https://dmogdx0jrul3u.cloudfront.net/1.7.9/lib', '/av');
					ZoomMtg.preLoadWasm();
					ZoomMtg.prepareJssdk();
					ZoomMtg.generateSignature({
						meetingNumber: data.meetingNumber,
						apiKey: data.apiKey,
						apiSecret: data.hash,
						role: 0,
						success: function (res) {
						  	console.log(res.result);
						  	ZoomMtg.init({
								leaveUrl: "/seminar",
								isSupportAV: true,
								screenShare: false,
								success: function() {
									ZoomMtg.join({
										signature: res.result,
										meetingNumber: data.meetingNumber,
										userName: data.userName,
										apiKey: data.apiKey,
										passWord: data.password,
										success: (success) => {
											console.log(success)
										},
										error: (error) => {
											console.log(error)
										}
									});
								}
							})
						},
					});
					
					
				})
				.fail(function(a,b,c){
					Swal.fire({
						type: 'error',
						title: 'Login Gagal: ' + c,
						text: a.responseText,
					});
				});

				
				

			});
		</script>
	</body>
	<script>
		$(document).ready(function(){
			var intervalID = setInterval(function(){
				$.get("/retoken")
				.done(function(data){

				});
			}, 270000);
		})
	</script>
	`

	return bufftemplate
}
