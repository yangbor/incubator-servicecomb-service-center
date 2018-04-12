/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"

	"github.com/apache/incubator-servicecomb-service-center/frontend/schema"
	"github.com/astaxie/beego"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	frontendIp := beego.AppConfig.String("frontend_host_ip")
	frontendPort := beego.AppConfig.DefaultInt("frontend_host_port", 30103)

	scIp := beego.AppConfig.DefaultString("httpadr", "127.0.0.1")
	scPort := beego.AppConfig.DefaultInt("httpport", 30100)

	// command line flags
	port := flag.Int("port", frontendPort, "port to serve on")
	dir := flag.String("directory", "app/", "directory of web files")

	flag.Parse()

	e := echo.New()
	// handle all requests by serving a file of the same name
	e.Static("/", *dir)

	e.Any("/testSchema/", schema.SchemaHandleFunc)

	// setup proxy for requests to service center
	scAddr := fmt.Sprintf("http://%s:%d", scIp, scPort)
	scUrl, err := url.Parse(scAddr)
	log.Printf("sc addr:%s", scAddr)
	if err != nil {
		log.Fatalf("Error parsing service center address:%s, err:%s", scAddr, err)
	}
	targets := []*middleware.ProxyTarget{
		{
			URL: scUrl,
		},
	}
	g := e.Group("/sc")
	balancer := middleware.NewRoundRobinBalancer(targets)
	pcfg := middleware.ProxyConfig{
		Balancer: balancer,
		Skipper:  middleware.DefaultSkipper,
		Rewrite: map[string]string{
			"/sc/*": "/$1",
		},
	}
	g.Use(middleware.ProxyWithConfig(pcfg))

	// run frontend web server
	log.Printf("Running on port %d\n", *port)
	addr := fmt.Sprintf("%s:%d", frontendIp, *port)

	// this call blocks -- the progam runs here forever
	log.Printf("Error: %s", e.Start(addr))
}
