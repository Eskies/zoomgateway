package controller

import (
	"database/sql"
	"strings"
	"zoomgateway/localtools"

	"github.com/huandu/go-sqlbuilder"

	"github.com/valyala/fasthttp"
)

func OnthespotPageController(dbConn *sql.DB, ctx *fasthttp.RequestCtx, pagesettings map[string]interface{}, zoomsettings map[string]interface{}) {
	if len(ctx.UserValue("nimmhs").(string)) > 3 {
		localtools.LogThisErrorWCode(ctx, 400, "YOU DONT HAVE ACCESS")
		return
	}

	output := oTSTemplate()
	output = strings.ReplaceAll(output, "$page_title", pagesettings["title"].(string))
	output = strings.ReplaceAll(output, "$page_description", pagesettings["description"].(string))
	output = strings.ReplaceAll(output, "$nim", ctx.UserValue("nimmhs").(string))
	output = strings.ReplaceAll(output, "$nama", ctx.UserValue("namamhs").(string))

	sqla := sqlbuilder.NewSelectBuilder()
	sqla.Select("id", "pembicara", "topik")
	sqla.From("sesi")
	a, b := sqla.Build()

	qresults, err := dbConn.Query(a, b...)
	if err != nil {
		localtools.LogThisError(ctx, err.Error())
		return
	}

	defer qresults.Close()
	options := ""
	for qresults.Next() {
		var id, pembicara, topik string
		qresults.Scan(
			&id,
			&pembicara,
			&topik,
		)
		options += `<option value="` + id + `">` + pembicara + " [" + topik + "]</option>"
	}
	output = strings.ReplaceAll(output, "$OPTIONSESI", options)

	ctx.WriteString(output)
	ctx.Response.Header.SetContentType("text/html; charset=UTF-8")
	ctx.SetConnectionClose()
	ctx.SetStatusCode(200)
	return
}

func oTSTemplate() string {
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
			.footer-page {
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
		<div class="container-fluid">
			<div class="row bg-info">
				<div class="col-12 text-left">
					<h5 class="text-white text-center">$page_title</h5>
				</div>
			</div>
			<div class="row bg-dark text-white">
				<div class="col-6 text-left">
					<p class="ml-1">ID: $nim<br />$nama</p>
				</div>
				<div class="col-6 text-right">
					<a href="/logout" class="text-danger"><button class="btn btn-danger btn-xs">Logout!</button></a>
				</div>
			</div>
			<div class="row justify-content-center align-items-center mt-5">
				<div class="col-md-6">
					<h3 class="text-center text-info">Pendaftaran On The Spot</h3>
					<div class="row">
						<div class="col-md-12">
							<div class="form-group">
								<label>ID / NIM</label>
								<input type="text" class="form-control" id="nimmhs">
							</div>
							<div class="form-group">
								<label>Nama Peserta</label>
								<input type="text" class="form-control" id="namamhs">
							</div>
							<div class="form-group">
								<label>Sesi Yang Dipilih</label>
								<select class="form-control" id="sesidipilih">
									$OPTIONSESI
								</select>
							</div>
							<div class="form-group">
								<button class="btn btn-success" type="button" id="cmd_daftar">DAFTAR!</button>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
		<div class="footer-page text-center">
			&copy; 2020. Developed by <a href="http://github.com/Eskies" target="_blank">NAK</a> @ <a href="http://stikom-bali.ac.id" target="_blank">stikom-bali.ac.id</a>
		</div>
	</body>
	<script>
		$(document).ready(function(){
			$("#cmd_daftar").click(function(){
				if (($("#nimmhs").val().length < 4) || ($("#namamhs").val().length < 4)) {
					Swal.fire({
						type: 'error',
						title: 'NIM dan NAMA minimal 4 karakter',
					});
					return false;
				}

				var datapost = {
					'nimmhs': $('#nimmhs').val(),
					'namamhs': $('#namamhs').val(),
					'sesi': $('#sesidipilih').val(),
				};
				$.post("/otsdaftar", datapost)
				.done(function(data){
					Swal.fire({
						type: 'success',
						title: 'DAFTAR BERHASIL!',
					});
					$('#nimmhs').val("");
					$('#namamhs').val("");
				})
				.fail(function(a,b,c){
					Swal.fire({
					  type: 'error',
					  title: 'Gagal Simpan: ' + c,
					  text: a.responseText,
					});
				});
			});

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

func OTSDaftar(dbConn *sql.DB, ctx *fasthttp.RequestCtx, pagesettings map[string]interface{}, zoomsettings map[string]interface{}) {
	if len(ctx.UserValue("nimmhs").(string)) > 3 {
		localtools.LogThisErrorWCode(ctx, 400, "YOU DONT HAVE ACCESS")
		return
	}

	ctx.SetStatusCode(200)

	nimmhs := string(ctx.FormValue("nimmhs"))
	namamhs := string(ctx.FormValue("namamhs"))
	sesi := string(ctx.FormValue("sesi"))

	sqla := sqlbuilder.NewSelectBuilder()
	sqla.Select("COUNT(*)")
	sqla.From("peserta")
	sqla.Where(sqla.E("id", nimmhs))
	a, b := sqla.Build()

	var nimada int
	err := dbConn.QueryRow(a, b...).Scan(&nimada)
	if err != nil {
		localtools.LogThisError(ctx, err.Error())
		return
	}

	if nimada == 0 {
		sqlb := sqlbuilder.NewInsertBuilder()
		sqlb.InsertInto("peserta")
		sqlb.Cols("id", "nama")
		sqlb.Values(nimmhs, namamhs)
		c, d := sqlb.Build()
		_, err := dbConn.Exec(c, d...)
		if err != nil {
			localtools.LogThisError(ctx, err.Error())
			return
		}
	}

	sqlc := sqlbuilder.NewInsertBuilder()
	sqlc.InsertInto("pesertapersesi")
	sqlc.Cols("peserta_id", "sesi_id")
	sqlc.Values(nimmhs, sesi)
	e, f := sqlc.Build()
	_, err = dbConn.Exec(e, f...)
	if err != nil {
		localtools.LogThisError(ctx, err.Error())
		return
	}

}
