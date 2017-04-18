/* Copyright 2017 Victor Penso

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>. */

package main

import (
  "strings"
  "regexp"
  "log"
  "io/ioutil"
  "os/exec"
  "github.com/prometheus/client_golang/prometheus"
) 

type NodesMetrics struct {
  alloc         float64
  comp          float64
  down          float64
  drain         float64
  err           float64        
  fail          float64
  idle          float64
  maint         float64
  mix           float64
  resv          float64
}

func NodesGetMetrics() *NodesMetrics {
  return ParseNodesMetrics(NodesData())
}

func ParseNodesMetrics(input []byte) *NodesMetrics {
  var nm NodesMetrics
  lines := strings.Split(string(input), "\n")
  for _, line := range lines {
    if strings.Contains(line, ",") {
      state := strings.Split(line,",")[1]
      alloc := regexp.MustCompile(`^alloc`)
      comp := regexp.MustCompile(`^comp`)
      down := regexp.MustCompile(`^down`)
      fail := regexp.MustCompile(`^fail`)
      err := regexp.MustCompile(`^err`)
      idle := regexp.MustCompile(`^idle`)
      maint := regexp.MustCompile(`^maint`)
      mix := regexp.MustCompile(`^mix`)
      resv:= regexp.MustCompile(`^resv`)
      switch {
      case alloc.MatchString(state) == true: nm.alloc++
      case comp.MatchString(state) == true: nm.comp++
      case down.MatchString(state) == true: nm.down++
      case fail.MatchString(state) == true: nm.fail++
      case err.MatchString(state) == true: nm.err++
      case idle.MatchString(state) == true: nm.idle++
      case maint.MatchString(state) == true: nm.maint++
      case mix.MatchString(state) == true: nm.mix++
      case resv.MatchString(state) == true: nm.resv++
      }
    }
  }
  return &nm
}

// Execute the squeue command and return its output
func NodesData() []byte {
  cmd := exec.Command("sinfo", "-h", "-o %n,%T", "| sort", "| uniq")
  stdout, err := cmd.StdoutPipe()
  if err != nil { log.Fatal(err) }
  if err := cmd.Start(); err != nil { log.Fatal(err) }
  out, _ := ioutil.ReadAll(stdout)
  if err := cmd.Wait(); err != nil { log.Fatal(err) }
  return out
}


/*
 * Implement the Prometheus Collector interface and feed the
 * Slurm scheduler metrics into it.
 * https://godoc.org/github.com/prometheus/client_golang/prometheus#Collector
 */

func NewNodesCollector() *NodesCollector {
  return &NodesCollector {
    alloc: prometheus.NewDesc("slurm_nodes_alloc","Allocated nodes",nil,nil),
    comp: prometheus.NewDesc("slurm_nodes_comp","Completing nodes",nil,nil),
    down: prometheus.NewDesc("slurm_nodes_down","Down nodes",nil,nil),
    drain: prometheus.NewDesc("slurm_nodes_drain","Drain nodes",nil,nil),
    err: prometheus.NewDesc("slurm_nodes_err","Error nodes",nil,nil),
    fail: prometheus.NewDesc("slurm_nodes_fail","Fail nodes",nil,nil),
    idle: prometheus.NewDesc("slurm_nodes_idle","Idle nodes",nil,nil),
    maint: prometheus.NewDesc("slurm_nodes_maint","Maint nodes",nil,nil),
    mix: prometheus.NewDesc("slurm_nodes_mix","Mix nodes",nil,nil),
    resv: prometheus.NewDesc("slurm_nodes_resv","Reserved nodes",nil,nil),
  } 
}

type NodesCollector struct {
  alloc      *prometheus.Desc   
  comp       *prometheus.Desc   
  down       *prometheus.Desc   
  drain      *prometheus.Desc   
  err        *prometheus.Desc           
  fail       *prometheus.Desc   
  idle       *prometheus.Desc   
  maint      *prometheus.Desc   
  mix        *prometheus.Desc   
  resv       *prometheus.Desc   
}


// Send all metric descriptions
func (nc *NodesCollector) Describe(ch chan<- *prometheus.Desc) {
  ch <- nc.alloc
  ch <- nc.comp
  ch <- nc.down
  ch <- nc.drain
  ch <- nc.err
  ch <- nc.fail
  ch <- nc.idle
  ch <- nc.maint
  ch <- nc.mix
  ch <- nc.resv
}
func (nc *NodesCollector) Collect(ch chan<- prometheus.Metric) {
  nm := NodesGetMetrics()
  ch <- prometheus.MustNewConstMetric(nc.alloc, prometheus.GaugeValue, nm.alloc)
  ch <- prometheus.MustNewConstMetric(nc.comp, prometheus.GaugeValue, nm.comp)
  ch <- prometheus.MustNewConstMetric(nc.down, prometheus.GaugeValue, nm.down)
  ch <- prometheus.MustNewConstMetric(nc.drain, prometheus.GaugeValue, nm.drain)
  ch <- prometheus.MustNewConstMetric(nc.err, prometheus.GaugeValue, nm.err)
  ch <- prometheus.MustNewConstMetric(nc.fail, prometheus.GaugeValue, nm.fail)
  ch <- prometheus.MustNewConstMetric(nc.idle, prometheus.GaugeValue, nm.idle)
  ch <- prometheus.MustNewConstMetric(nc.maint, prometheus.GaugeValue, nm.maint)
  ch <- prometheus.MustNewConstMetric(nc.mix, prometheus.GaugeValue, nm.mix)
  ch <- prometheus.MustNewConstMetric(nc.resv, prometheus.GaugeValue, nm.resv)
}


