/*
 * skogul, influxdb writer
 *
 * Copyright (c) 2019 Telenor Norge AS
 * Author(s):
 *  - Kristian Lyngstøl <kly@kly.no>
 *
 * This library is free software; you can redistribute it and/or
 * modify it under the terms of the GNU Lesser General Public
 * License as published by the Free Software Foundation; either
 * version 2.1 of the License, or (at your option) any later version.
 *
 * This library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public
 * License along with this library; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA
 * 02110-1301  USA
 */

package senders

import (
	"bytes"
	"fmt"
	"github.com/KristianLyng/skogul/pkg"
	"log"
	"net/http"
	"sync"
	"time"
)

type InfluxDB struct {
	URL         string
	Measurement string
	client      *http.Client
	mux         sync.Mutex
}

func (idb *InfluxDB) Send(c *skogul.Container) error {
	var buffer bytes.Buffer
	for _, m := range c.Metrics {
		fmt.Fprintf(&buffer, "%s", idb.Measurement)
		for key, value := range m.Metadata {
			fmt.Fprintf(&buffer, ",%s=%#v", key, value)
		}
		fmt.Fprintf(&buffer, " ")
		comma := ""
		for key, value := range m.Data {
			fmt.Fprintf(&buffer, "%s%s=%#v", comma, key, value)
			comma = ","
		}
		fmt.Fprintf(&buffer, " %d\n", m.Time.UnixNano())
	}
	if idb.client == nil {
		idb.mux.Lock()
		// Recheck after acquiring lock
		if idb.client == nil {
			idb.client = &http.Client{Timeout: 5 * time.Second}
		}
		idb.mux.Unlock()
	}
	resp, err := idb.client.Post(idb.URL, "text/plain", &buffer)
	if err != nil {
		log.Print(err)
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Print(resp)
		return skogul.Gerror{"Bad response code from InfluxDB"}
	}
	return nil
}

//req, err := http.NewRequest("POST", "http://127.0.0.1:8086/write?db=test", &buffer)
