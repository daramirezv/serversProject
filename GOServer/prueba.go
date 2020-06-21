package main

import (
	"fmt"
	"log"
	"net/http"
	"database/sql"
	"strconv"
	"io/ioutil"
	"encoding/json"
	"time"
	"strings"
	"github.com/gocolly/colly"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	_ "github.com/lib/pq"
)

//Funcion que retorna un JSON con las paginas antes revisadas
func Consultas(ctx *fasthttp.RequestCtx) {

	//Conexion a la base de datos
	db, err := sql.Open("postgres", "postgresql://maxroach@localhost:26257/bank?sslmode=disable")
    if err != nil {
        log.Fatal("error connecting to the database: ", err)
	}

	//Seleccion de todos los valores que se encuentren en la base de datos de consultas
	rows, err := db.Query("SELECT id FROM consultas")
    if err != nil {
        log.Fatal(err)
	}
	
	defer rows.Close()

	//Se crea el string donde se van a guardar todos los valores antes consultados
	var id string
	var respuesta string = "{\"items\":["
	var existen = false

	//Se hacen iteraciones por cada valor que se encuentre en la base de datos. Se va construyendo el json respuesta
    for rows.Next() {
		existen = true
        if err := rows.Scan(&id); err != nil {
            log.Fatal(err)
		}
		respuesta = respuesta + "{\"name\":\"" + id + "\"},"
	}
	if existen {
		respuesta = strings.TrimRight(respuesta, ",")
	}
	respuesta = respuesta + "]}"
	
	//Se devuelve el json respuesta a quien haya accedido al endpoint
	ctx.Response.Header.SetCanonical([]byte("Content-Type"), []byte("application/json"))
	ctx.Response.Header.SetCanonical([]byte("Access-Control-Allow-Methods"), []byte("HEAD,GET,POST,PUT,DELETE,OPTIONS"))
	ctx.Response.Header.SetCanonical([]byte("Access-Control-Allow-Origin"), []byte("*"))
	ctx.Response.Header.SetCanonical([]byte("Access-Control-Allow-Credentials"), []byte("true"))
	fmt.Fprint(ctx, respuesta)
}

//Funcion que consulta al API toda la informacion relacionada con la pagina buscada
func Dominio(ctx *fasthttp.RequestCtx) {

	var logo string = "No tiene un enlace al logo de la pagina"
	var tituloPagina string = "No tiene un titulo en la pagina"

	//Peticiones al API de colly (scrapper/crawler que revisa el logo y el titulo de la pagina)
	c := colly.NewCollector()
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 1})
	c.OnHTML("link[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		rel := e.Attr("rel")
		if rel == "shortcut icon" {
			logo = link
		}
	})
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})
	c.Visit("https://www." + ctx.UserValue("host").(string))

	d := colly.NewCollector()
	d.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 1})
	d.OnHTML("title", func(e *colly.HTMLElement) {
		tituloPagina = e.Text
	})
	d.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})
	d.Visit("https://www." + ctx.UserValue("host").(string))

	//String donde se va a guardar el json de respuesta final
	var respuestaJson string = "{\"servers\":["

	//Conexion a la base de datos
	db, err := sql.Open("postgres", "postgresql://maxroach@localhost:26257/bank?sslmode=disable")
    if err != nil {
        log.Fatal("error connecting to the database: ", err)
	}

	//API de ssllabs para buscar los servidores de la pagina
	var url string = "https://api.ssllabs.com/api/v3/analyze?host=" + ctx.UserValue("host").(string)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	//Query que busca si antes ya se ha buscado dicha pagina
	rows, err := db.Query("SELECT id, numero FROM consultas WHERE id = '" + ctx.UserValue("host").(string) + "'")
    if err != nil {
        log.Fatal(err)
    }
	defer rows.Close()

	//Si la página se ha consultado antes, actualiza el número de veces buscada o se crea un nuevo elemento si no se ha consultado
	if rows.Next() {
		var id string
		var numero int
        if err := rows.Scan(&id, &numero); err != nil {
            log.Fatal(err)
		}
		if _, err := db.Exec(
			"UPDATE consultas SET numero = " + strconv.Itoa(numero+1) + " WHERE id = '" + ctx.UserValue("host").(string) + "'"); err != nil {
			log.Fatal(err)
		}
	} else {
		if _, err2 := db.Exec(
			"INSERT INTO consultas (id, numero) VALUES ('" + ctx.UserValue("host").(string) + "', 1)"); err2 != nil {
			log.Fatal(err2)
		}
	}

	//Se toma el json consultado del API de ssllabs para ser procesado.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
        log.Fatal(err)
	}

	in := []byte(string(body))
	var objeto map[string]interface{}
	err2 := json.Unmarshal(in, &objeto)
	if err2 != nil {
        fmt.Println(err2)
	}

	var estatus string = objeto["status"].(string)
	
	//Se valida que la pagina tenga servidores
	if objeto["endpoints"] == nil {
		ctx.Response.Header.SetCanonical([]byte("Content-Type"), []byte("application/json"))
		ctx.Response.Header.SetCanonical([]byte("Access-Control-Allow-Methods"), []byte("HEAD,GET,POST,PUT,DELETE,OPTIONS"))
		ctx.Response.Header.SetCanonical([]byte("Access-Control-Allow-Origin"), []byte("*"))
		ctx.Response.Header.SetCanonical([]byte("Access-Control-Allow-Credentials"), []byte("true"))
		fmt.Fprint(ctx, "{\"message\":\"The site you searched has no servers.\"}")
	} else {
		//En dichas variables se guarda el mejor rating nuevo y antiguo para las paginas revisadas
		//D es el peor rating posible y A+ el mejor rating posible
		var cambioHaceUnaHora bool = false;
		var hayValorViejo bool = false;
		var viejoMayor string = "A+"
		var nuevoMayor string = "A+"
		//Se hace una iteracion por todos los servidores con base en la informacion del API
		for k, v := range objeto["endpoints"].([]interface {}) {
			//IP traido del API
			var ipAddressPeticion string = v.(map[string]interface{})["ipAddress"].(string)
			//Rating traido del API
			var gradePeticion string = v.(map[string]interface{})["grade"].(string)

			respuestaJson = respuestaJson + "{\"address\":\""+ ipAddressPeticion +"\", \"ssl_grade\":\"" + gradePeticion + "\"},"

			println(k)

			//Query que consulta todos los servidores guardados en la BD con la relacion a la pagina que busco el usuario
			rowsBD, err := db.Query("SELECT id, creation_time, ipAddress, grade FROM servidores WHERE busqueda = '" + ctx.UserValue("host").(string) + "' AND ipAddress = '" + ipAddressPeticion + "'")
			if err != nil {
				log.Fatal(err)
			}
			defer rowsBD.Close()
	
			//Se revisa si dicho servidor ya se busco antes
			if rowsBD.Next() {
				hayValorViejo = true
				var fecha64Actual int64 = time.Now().Unix()
				var fechaActual int = int(fecha64Actual)
			
				var id string
				var fecha int
				var ipAddress string
				var grade string
				

				if err := rowsBD.Scan(&id, &fecha, &ipAddress, &grade); err != nil {
					log.Fatal(err)
				}

				//Se actualizan las variables del servidor con el peor rating de ahora y de los servidores antes consultados
				if nuevoMayor == "A+" && gradePeticion != "A+" {
					nuevoMayor = gradePeticion
				} else if gradePeticion > nuevoMayor {
					nuevoMayor = gradePeticion
				}

				if viejoMayor == "A+" && grade != "A+" {
					viejoMayor = grade
				} else if grade > nuevoMayor {
					viejoMayor = grade
				}

				//Si el rating o la IP de un servidor cambio hace una hora o mas, se notifica dicho cambio en el json. 
				//Si cambio en menos de una hora, no se notifica que cambio, pero se modifica la base de datos.
				if (fechaActual - fecha) < 3600 {
					if _, err := db.Exec(
						"UPDATE servidores SET creation_time = " + strconv.Itoa(fechaActual) + ", ipAddress = '"+ipAddressPeticion+"', grade = '"+gradePeticion+"' WHERE id = '" + id + "'"); err != nil {
						log.Fatal(err)
					}
				} else {
					if ipAddress != ipAddressPeticion || grade != gradePeticion {
						cambioHaceUnaHora = true
						if _, err := db.Exec(
							"UPDATE servidores SET creation_time = " + strconv.Itoa(fechaActual) + ", ipAddress = '"+ipAddressPeticion+"', grade = '"+gradePeticion+"' WHERE id = '" + id + "'"); err != nil {
							log.Fatal(err)
						}
					}
				}
			//Este es el caso que un servidor no se haya revisado antes.
			} else {
				if nuevoMayor == "A+" && gradePeticion != "A+" {
					nuevoMayor = gradePeticion
				} else if gradePeticion > nuevoMayor {
					nuevoMayor = gradePeticion
				}
				var fecha64 int64 = time.Now().Unix()
				var fecha int = int(fecha64)
				if _, err2 := db.Exec(
					"INSERT INTO servidores (creation_time, busqueda, ipAddress, grade) VALUES (" + strconv.Itoa(fecha) + ", '"+ ctx.UserValue("host").(string) +"','"+ipAddressPeticion+"','"+gradePeticion+"')"); err2 != nil {
					log.Fatal(err2)
				}
			}
		}

		//Se va completando el json final.
		respuestaJson = strings.TrimRight(respuestaJson, ",")
		respuestaJson = respuestaJson + "],\"servers_changed\":"+strconv.FormatBool(cambioHaceUnaHora)
		respuestaJson = respuestaJson + ",\"ssl_grade\":\""+nuevoMayor+"\""
		if hayValorViejo {
			respuestaJson = respuestaJson + ",\"previous_ssl_grade\":\""+viejoMayor+"\""
		} else {
			respuestaJson = respuestaJson + ",\"previous_ssl_grade\":\"No Hay Anterior\""
		}

		respuestaJson = respuestaJson + ",\"logo\":\""+logo+"\""
		respuestaJson = respuestaJson + ",\"title\":\""+tituloPagina+"\""

		if estatus == "READY" {
			respuestaJson = respuestaJson + ",\"is_down\":false}"
		} else {
			respuestaJson = respuestaJson + ",\"is_down\":true}"
		}

		ctx.Response.Header.SetCanonical([]byte("Content-Type"), []byte("application/json"))
		ctx.Response.Header.SetCanonical([]byte("Access-Control-Allow-Methods"), []byte("HEAD,GET,POST,PUT,DELETE,OPTIONS"))
		ctx.Response.Header.SetCanonical([]byte("Access-Control-Allow-Origin"), []byte("*"))
		ctx.Response.Header.SetCanonical([]byte("Access-Control-Allow-Credentials"), []byte("true"))
		fmt.Fprint(ctx, respuestaJson)
	}
}

//Funcion donde se especifica que funcion corresponde a cada endpoint.
func main() {
	router := fasthttprouter.New()

	db, err := sql.Open("postgres", "postgresql://maxroach@localhost:26257/bank?sslmode=disable")
    if err != nil {
        log.Fatal("error connecting to the database: ", err)
	}

	//Los metodos de abajo son para borrar las bases de datos al inicio del servidor.

	// if _, err := db.Exec(
	// 	"DROP TABLE consultas"); err != nil {
    //     log.Fatal(err)
	// }

	// if _, err := db.Exec(
	// 	"DROP TABLE servidores"); err != nil {
    //     log.Fatal(err)
	// }
	
    if _, err := db.Exec(
		"CREATE TABLE IF NOT EXISTS consultas (id STRING PRIMARY KEY, numero INT)"); err != nil {
        log.Fatal(err)
	}
	
	if _, err := db.Exec(
		"CREATE TABLE IF NOT EXISTS servidores (id UUID NOT NULL DEFAULT gen_random_uuid(), creation_time INT, busqueda STRING, ipAddress STRING, grade STRING)"); err != nil {
        log.Fatal(err)
	}

	//Endpoints y que metodos responden a cada endpoint.
	router.GET("/consultas", Consultas)
	router.GET("/dominio/:host", Dominio)

	//Puerto donde se va a ejecutar el API.
	log.Fatal(fasthttp.ListenAndServe(":8085", router.Handler))
}