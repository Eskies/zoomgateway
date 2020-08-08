package controller

import (
	"database/sql"
	"strings"
	"zoomgateway/localtools"

	"github.com/huandu/go-sqlbuilder"

	"github.com/valyala/fasthttp"
)

func SeminarPageController(dbConn *sql.DB, ctx *fasthttp.RequestCtx, pagesettings map[string]interface{}, zoomsettings map[string]interface{}) {
	output := SeminarTemplate()

	output = strings.ReplaceAll(output, "$page_title", pagesettings["title"].(string))
	output = strings.ReplaceAll(output, "$page_description", pagesettings["description"].(string))
	output = strings.ReplaceAll(output, "$nim", ctx.UserValue("nimmhs").(string))
	output = strings.ReplaceAll(output, "$nama", ctx.UserValue("namamhs").(string))

	//build list seminar
	var daftarSeminar string = ``
	templateDaftar := `
		<div class="col-sm-12 mb-2">
			<div class="card">
				<div class="card-header bg-info text-white text-center">
					<h5 class="card-title">Topik: $topik</h5>
					<h6 class="card-subtitle text-white">Pembicara: $pembicara</h6>
				</div>
				<div class="card-body">
					<p class="card-text">Seminar dilaksanakan tanggal <b>$tanggal_seminar</b> dari waktu <b>$waktu_mulai WITA</b> hingga <b>$waktu_selesai WITA</b></p>
					<p class="card-text">Join anda terakhir: $last_join</p>
					<a href="$link_seminar" class="btn btn-success btn-xs">
						Join Seminar
					</a>
				</div>
			</div>
		</div>
	`

	sqlB := sqlbuilder.NewSelectBuilder()
	sqlB.Select("sesi.id", "sesi.topik", "sesi.pembicara", `DATE_FORMAT(sesi.tanggal, "%d-%b-%Y")`, "sesi.waktumulai", "sesi.waktuselesai", `DATE_FORMAT(pesertapersesi.waktulogin, "jam %H:%i:%s tanggal %d-%b-%Y")`)
	sqlB.From("pesertapersesi")
	sqlB.Join("sesi", "sesi.id = pesertapersesi.sesi_id")
	sqlB.Where(sqlB.E("peserta_id", ctx.UserValue("nimmhs").(string)))

	a, b := sqlB.Build()
	qresults, err := dbConn.Query(a, b...)
	if err != nil {
		localtools.LogThisError(ctx, err.Error())
		return
	}

	defer qresults.Close()
	for qresults.Next() {
		var id, topik, pembicara, tanggal, mulai, selesai string
		var waktulogin sql.NullString

		var err = qresults.Scan(
			&id,
			&topik,
			&pembicara,
			&tanggal,
			&mulai,
			&selesai,
			&waktulogin,
		)
		if err != nil {
			localtools.LogThisError(ctx, err.Error())
			return
		}

		buffSeminarItem := strings.ReplaceAll(templateDaftar, "$topik", topik)
		buffSeminarItem = strings.ReplaceAll(buffSeminarItem, "$pembicara", pembicara)
		buffSeminarItem = strings.ReplaceAll(buffSeminarItem, "$tanggal_seminar", tanggal)
		buffSeminarItem = strings.ReplaceAll(buffSeminarItem, "$waktu_mulai", mulai)
		buffSeminarItem = strings.ReplaceAll(buffSeminarItem, "$waktu_selesai", selesai)
		buffSeminarItem = strings.ReplaceAll(buffSeminarItem, "$link_seminar", "/joinseminar/"+id)
		if waktulogin.Valid {
			buffSeminarItem = strings.ReplaceAll(buffSeminarItem, "$last_join", waktulogin.String)
		} else {
			buffSeminarItem = strings.ReplaceAll(buffSeminarItem, "$last_join", "(belum pernah join)")
		}

		daftarSeminar += buffSeminarItem
	}

	output = strings.ReplaceAll(output, "$list_seminar", daftarSeminar)
	ctx.WriteString(output)

	ctx.Response.Header.SetContentType("text/html; charset=UTF-8")
	ctx.SetConnectionClose()
	ctx.SetStatusCode(200)

	return
}

func SeminarTemplate() string {
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
					<h3 class="text-center text-info">Seminar yang berhak anda ikuti</h3>
					<div class="row">
						$list_seminar
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
