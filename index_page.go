package main

import (
	"database/sql"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/valyala/fasthttp"
)

func IndexController(dbConn *sql.DB, ctx *fasthttp.RequestCtx, pagesettings map[string]interface{}, zoomsettings map[string]interface{}) {
	tokenstring := string(ctx.Request.Header.Cookie("auth"))
	if len(tokenstring) > 0 {
		token, _ := jwt.Parse(tokenstring, func(token *jwt.Token) (interface{}, error) {
			return jwtkey, nil
		})

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if string(ctx.Host()) == claims["aud"] {
				nimmhs := claims["nim"].(string)
				sscodeReq := claims["sscode"].(string)
				if x, found := poolcache.Get("psr-" + nimmhs); found {
					if sscodeReq == x.(string) {
						ctx.Redirect("/seminar", 307)
						return
					}
				}
			}
		}
	}

	output := IndexTemplate()

	output = strings.ReplaceAll(output, "$page_title", pagesettings["title"].(string))
	output = strings.ReplaceAll(output, "$page_description", pagesettings["description"].(string))

	ctx.Response.Header.SetContentType("text/html; charset=UTF-8")
	ctx.WriteString(output)
	ctx.SetConnectionClose()
	ctx.SetStatusCode(200)

	return
}

func IndexTemplate() string {
	var bufftemplate string
	bufftemplate = `
	<!DOCTYPE html>
	<html lang="en">

	<head>
		<meta charset="utf-8">
		<meta http-equiv="X-UA-Compatible" content="IE=edge">
		<title>$page_title</title>
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<meta name="description" content="$page_description">
		<meta name="author" content="Ngurah Ady Kusuma (github.com/Eskies) (email: ady_kusuma@stikom-bali.ac.id">
		
		<link href="//maxcdn.bootstrapcdn.com/bootstrap/4.1.1/css/bootstrap.min.css" rel="stylesheet" id="bootstrap-css">
		<script src="//cdnjs.cloudflare.com/ajax/libs/jquery/3.2.1/jquery.min.js"></script>
		<script src="//maxcdn.bootstrapcdn.com/bootstrap/4.1.1/js/bootstrap.min.js"></script>
		
		<script src="https://cdn.jsdelivr.net/npm/sweetalert2@9"></script>
		<style>
			html {
				position: relative;
				min-height: 100%;
			}
			body {
				margin-bottom: 60px; /* Margin bottom by footer height */
			}
			.footer {
				position: absolute;
				bottom: 0;
				width: 100%;
				height: 60px; /* Set the fixed height of the footer here */
				line-height: 60px; /* Vertically center the text there */
				background-color: #f5f5f5;
			}
		</style>
	</head>
	<body>
		<div id="login">
			<h3 class="text-center text-white pt-5">Login form</h3>
			<div class="container">
				<div id="login-row" class="row justify-content-center align-items-center">
					<div id="login-column" class="col-md-6">
						<div id="login-box" class="col-md-12">
							<form class="form">
								<h3 class="text-center text-info">$page_title</h3>
								<p class="text-center text-muted">$page_description</p>

								<div class="form-group" style="margin-top: 30px;">
									<label for="username" class="text-info">ID / NIM</label><br>
									<input type="text" name="username" id="username" class="form-control">
									<p class="text-muted" style="font-size: 10px;">Perhatian, anda hanya boleh login sebanyak satu kali. Double login akan diblokir. Masukan ID yang anda terima untuk mengikuti seminar ini</p>
								</div>
								<div class="form-group">
									<button type="button" class="btn btn-block btn-info btn-md" id="login_btn">Masuk Seminar</button>
								</div>
							</form>
						</div>
					</div>
				</div>
			</div>
		</div>
		<div class="footer text-center">
			&copy; 2020. Developed by <a href="http://github.com/Eskies" target="_blank">NAK</a> @ <a href="http://stikom-bali.ac.id" target="_blank">stikom-bali.ac.id</a>
		</div>
	</body>
	<script>
		$(document).ready(function(){
			$('#login_btn').click(function(){
			if (($('#username').val().trim().length > 0)) {
				var datapost = {
					'username': $('#username').val(),
				};
				$.post("/login", datapost)
				.done(function(data){
					Swal.fire({
						type: 'success',
						title: 'Login berhasil!',
					});
					window.location.replace("/seminar");
				})
				.fail(function(a,b,c){
					Swal.fire({
						type: 'error',
						title: 'Login Gagal: ' + c,
						text: a.responseText,
					});
				});
			} else {
				Swal.fire({type: 'error', title: 'Login',text: "Mohon inputkan ID anda"});
			}
			});
		});
	</script>
	`

	return bufftemplate
}
