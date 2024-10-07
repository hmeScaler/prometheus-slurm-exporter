/* Copyright 2017 Victor Penso, Matteo Dessalvi

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
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"log"
	"os/exec"
//	"regexp"
//	"sort"
//	"strconv"
	"strings"
)

type NodesMetrics struct {
    alloc     float64
    allocNodes map[string]bool
    comp      float64
    compNodes  map[string]bool
    down      float64
    downNodes  map[string]bool
    drain     float64
    drainNodes map[string]bool
    err       float64
    errNodes   map[string]bool
    fail      float64
    failNodes  map[string]bool
    idle      float64
    idleNodes  map[string]bool
    maint     float64
    maintNodes map[string]bool
    mix       float64
    mixNodes   map[string]bool
    resv      float64
    resvNodes  map[string]bool
}


func NodesGetMetrics() *NodesMetrics {
    output := NodesData()
    return ParseNodesMetrics(output)
}

func RemoveDuplicates(s []string) []string {
	m := map[string]bool{}
	t := []string{}

	// Walk through the slice 's' and for each value we haven't seen so far, append it to 't'.
	for _, v := range s {
		if _, seen := m[v]; !seen {
			if len(v) > 0 {
				t = append(t, v)
				m[v] = true
			}
		}
	}

	return t
}

func ParseNodesMetrics(input []byte) *NodesMetrics {
    nm := &NodesMetrics{
        allocNodes: make(map[string]bool),
        compNodes:  make(map[string]bool),
        downNodes:  make(map[string]bool),
        drainNodes: make(map[string]bool),
        errNodes:   make(map[string]bool),
        failNodes:  make(map[string]bool),
        idleNodes:  make(map[string]bool),
        maintNodes: make(map[string]bool),
        mixNodes:   make(map[string]bool),
        resvNodes:  make(map[string]bool),
    }
    lines := strings.Split(string(input), "\n")
    for _, line := range lines {
        if strings.Contains(line, ",") {
            split := strings.Split(line, ",")
            if len(split) < 2 {
                continue
            }
            nodeName := strings.TrimSpace(split[0])
            state := strings.TrimSpace(split[1])

	    switch {
            case strings.HasPrefix(state, "alloc"):
                if !nm.allocNodes[nodeName] {
                    nm.alloc++
                    nm.allocNodes[nodeName] = true
                }
            case strings.HasPrefix(state, "comp"):
                if !nm.compNodes[nodeName] {
                    nm.comp++
                    nm.compNodes[nodeName] = true
                }
            case strings.HasPrefix(state, "down"):
                if !nm.downNodes[nodeName] {
                    nm.down++
                    nm.downNodes[nodeName] = true
                }
            case strings.HasPrefix(state, "drain"):
                if !nm.drainNodes[nodeName] {
                    nm.drain++
                    nm.drainNodes[nodeName] = true
                }
            case strings.HasPrefix(state, "err"):
                if !nm.errNodes[nodeName] {
                    nm.err++
                    nm.errNodes[nodeName] = true
                }
            case strings.HasPrefix(state, "fail"):
                if !nm.failNodes[nodeName] {
                    nm.fail++
                    nm.failNodes[nodeName] = true
                }
            case strings.HasPrefix(state, "idle"):
                if !nm.idleNodes[nodeName] {
                    nm.idle++
                    nm.idleNodes[nodeName] = true
                }
            case strings.HasPrefix(state, "maint"):
                if !nm.maintNodes[nodeName] {
                    nm.maint++
                    nm.maintNodes[nodeName] = true
                }
            case strings.HasPrefix(state, "mix"):
                if !nm.mixNodes[nodeName] {
                    nm.mix++
                    nm.mixNodes[nodeName] = true
                }
            case strings.HasPrefix(state, "resv"):
                if !nm.resvNodes[nodeName] {
                    nm.resv++
                    nm.resvNodes[nodeName] = true
                }
            }
        }
    }
    return nm
}

// Execute the sinfo command and return its output
func NodesData() []byte {
	//cmd := exec.Command("/cm/shared/apps/slurm/current/bin/sinfo", "-h", "-o %D,%T,%N")
	cmd := exec.Command("/cm/shared/apps/slurm/current/bin/sinfo", "-h", "-N", "-o %N,%T")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	out, _ := ioutil.ReadAll(stdout)
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	//log.Printf("NodesData output: %s", string(out))  // Ajoutez cette ligne
	return out
}

/*
 * Implement the Prometheus Collector interface and feed the
 * Slurm scheduler metrics into it.
 * https://godoc.org/github.com/prometheus/client_golang/prometheus#Collector
 */

func NewNodesCollector() *NodesCollector {
    return &NodesCollector{
        alloc:     prometheus.NewDesc("slurm_nodes_alloc", "Allocated nodes", nil, nil),
        allocNodes: prometheus.NewDesc("slurm_nodes_alloc_names", "Names of allocated nodes", []string{"node"}, nil),
        comp:      prometheus.NewDesc("slurm_nodes_comp", "Completing nodes", nil, nil),
        compNodes:  prometheus.NewDesc("slurm_nodes_comp_names", "Names of completing nodes", []string{"node"}, nil),
        down:      prometheus.NewDesc("slurm_nodes_down", "Down nodes", nil, nil),
        downNodes:  prometheus.NewDesc("slurm_nodes_down_names", "Names of down nodes", []string{"node"}, nil),
        drain:     prometheus.NewDesc("slurm_nodes_drain", "Drain nodes", nil, nil),
        drainNodes: prometheus.NewDesc("slurm_nodes_drain_names", "Names of drain nodes", []string{"node"}, nil),
        err:       prometheus.NewDesc("slurm_nodes_err", "Error nodes", nil, nil),
        errNodes:   prometheus.NewDesc("slurm_nodes_err_names", "Names of error nodes", []string{"node"}, nil),
        fail:      prometheus.NewDesc("slurm_nodes_fail", "Fail nodes", nil, nil),
        failNodes:  prometheus.NewDesc("slurm_nodes_fail_names", "Names of fail nodes", []string{"node"}, nil),
        idle:      prometheus.NewDesc("slurm_nodes_idle", "Idle nodes", nil, nil),
        idleNodes:  prometheus.NewDesc("slurm_nodes_idle_names", "Names of idle nodes", []string{"node"}, nil),
        maint:     prometheus.NewDesc("slurm_nodes_maint", "Maint nodes", nil, nil),
        maintNodes: prometheus.NewDesc("slurm_nodes_maint_names", "Names of maint nodes", []string{"node"}, nil),
        mix:       prometheus.NewDesc("slurm_nodes_mix", "Mix nodes", nil, nil),
        mixNodes:   prometheus.NewDesc("slurm_nodes_mix_names", "Names of mix nodes", []string{"node"}, nil),
        resv:      prometheus.NewDesc("slurm_nodes_resv", "Reserved nodes", nil, nil),
        resvNodes:  prometheus.NewDesc("slurm_nodes_resv_names", "Names of reserved nodes", []string{"node"}, nil),
    }
}

type NodesCollector struct {
    alloc     *prometheus.Desc
    allocNodes *prometheus.Desc
    comp      *prometheus.Desc
    compNodes  *prometheus.Desc
    down      *prometheus.Desc
    downNodes  *prometheus.Desc
    drain     *prometheus.Desc
    drainNodes *prometheus.Desc
    err       *prometheus.Desc
    errNodes   *prometheus.Desc
    fail      *prometheus.Desc
    failNodes  *prometheus.Desc
    idle      *prometheus.Desc
    idleNodes  *prometheus.Desc
    maint     *prometheus.Desc
    maintNodes *prometheus.Desc
    mix       *prometheus.Desc
    mixNodes   *prometheus.Desc
    resv      *prometheus.Desc
    resvNodes  *prometheus.Desc
}


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
    for node := range nm.allocNodes {
        ch <- prometheus.MustNewConstMetric(nc.allocNodes, prometheus.GaugeValue, 1, node)
    }
    ch <- prometheus.MustNewConstMetric(nc.comp, prometheus.GaugeValue, nm.comp)
    for node := range nm.compNodes {
        ch <- prometheus.MustNewConstMetric(nc.compNodes, prometheus.GaugeValue, 1, node)
    }
    ch <- prometheus.MustNewConstMetric(nc.down, prometheus.GaugeValue, nm.down)
    for node := range nm.downNodes {
        ch <- prometheus.MustNewConstMetric(nc.downNodes, prometheus.GaugeValue, 1, node)
    }
    ch <- prometheus.MustNewConstMetric(nc.drain, prometheus.GaugeValue, nm.drain)
    for node := range nm.drainNodes {
        ch <- prometheus.MustNewConstMetric(nc.drainNodes, prometheus.GaugeValue, 1, node)
    }
    ch <- prometheus.MustNewConstMetric(nc.err, prometheus.GaugeValue, nm.err)
    for node := range nm.errNodes {
        ch <- prometheus.MustNewConstMetric(nc.errNodes, prometheus.GaugeValue, 1, node)
    }
    ch <- prometheus.MustNewConstMetric(nc.fail, prometheus.GaugeValue, nm.fail)
    for node := range nm.failNodes {
        ch <- prometheus.MustNewConstMetric(nc.failNodes, prometheus.GaugeValue, 1, node)
    }
    ch <- prometheus.MustNewConstMetric(nc.idle, prometheus.GaugeValue, nm.idle)
    for node := range nm.idleNodes {
        ch <- prometheus.MustNewConstMetric(nc.idleNodes, prometheus.GaugeValue, 1, node)
    }
    ch <- prometheus.MustNewConstMetric(nc.maint, prometheus.GaugeValue, nm.maint)
    for node := range nm.maintNodes {
        ch <- prometheus.MustNewConstMetric(nc.maintNodes, prometheus.GaugeValue, 1, node)
    }
    ch <- prometheus.MustNewConstMetric(nc.mix, prometheus.GaugeValue, nm.mix)
    for node := range nm.mixNodes {
        ch <- prometheus.MustNewConstMetric(nc.mixNodes, prometheus.GaugeValue, 1, node)
    }
    ch <- prometheus.MustNewConstMetric(nc.resv, prometheus.GaugeValue, nm.resv)
    for node := range nm.resvNodes {
        ch <- prometheus.MustNewConstMetric(nc.resvNodes, prometheus.GaugeValue, 1, node)
    }
}
