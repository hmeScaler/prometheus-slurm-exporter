/* Copyright 2020 Victor Penso

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
        "io/ioutil"
        "os/exec"
        "log"
        "strings"
        "strconv"
        "regexp"
        "github.com/prometheus/client_golang/prometheus"
)

func UsersData() []byte {
        cmd := exec.Command("/cm/shared/apps/slurm/current/bin/squeue","-a","-r","-h","-o %A|%u|%T|%C|%b")
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
	return out
}

type UserJobMetrics struct {
        pending float64
        pending_cpus float64
	pending_gpus float64
        running float64
        running_cpus float64
	running_gpus  float64
        suspended float64
}

func ParseUsersMetrics(input []byte) map[string]*UserJobMetrics {
        users := make(map[string]*UserJobMetrics)
        lines := strings.Split(string(input), "\n")
        for _, line := range lines {
                if strings.Contains(line,"|") {
			parts := strings.Split(line, "|")
                        user := strings.Split(line,"|")[1]
                        _,key := users[user]
                        if !key {
                                users[user] = &UserJobMetrics{0,0,0,0,0,0,0}
                        }
                        state := strings.Split(line,"|")[2]
                        state = strings.ToLower(state)
                        cpus,_ := strconv.ParseFloat(strings.Split(line,"|")[3],64)
			gres := parts[4]
                        pending := regexp.MustCompile(`^pending`)
                        running := regexp.MustCompile(`^running`)
                        suspended := regexp.MustCompile(`^suspended`)

			gpus := 0.0
			if gres != "N/A" {
				gresParts := strings.Split(gres, ",")
				for _, g := range gresParts {
					if strings.Contains(g, "gpu:") {
						gpuParts := strings.Split(g, ":")
						if len(gpuParts) == 3 {
							numGPUs, err := strconv.ParseFloat(gpuParts[2], 64)
							if err != nil {
								log.Printf("Erreur de parsing du nombre de GPU : %v", err)
							}
							gpus += numGPUs
						}
					}
				}
			}

			//log.Printf("User: %s, State: %s, CPUs: %f, GPUs: %f", user, state, cpus, gpus) // Ajout du log

                        switch {
                        case pending.MatchString(state) == true:
                                users[user].pending++
                                users[user].pending_cpus += cpus
				users[user].pending_gpus += gpus
                        case running.MatchString(state) == true:
                                users[user].running++
                                users[user].running_cpus += cpus
				users[user].running_gpus += gpus
                        case suspended.MatchString(state) == true:
                                users[user].suspended++
                        }
                }
        }
        return users
}

type UsersCollector struct {
        pending *prometheus.Desc
        pending_cpus *prometheus.Desc
	pending_gpus  *prometheus.Desc
        running *prometheus.Desc
        running_cpus *prometheus.Desc
	running_gpus *prometheus.Desc
        suspended *prometheus.Desc
}

func NewUsersCollector() *UsersCollector {
        labels := []string{"user"}
        return &UsersCollector {
                pending: prometheus.NewDesc("slurm_user_jobs_pending", "Pending jobs for user", labels, nil), 
                pending_cpus: prometheus.NewDesc("slurm_user_cpus_pending", "Pending jobs for user", labels, nil), 
		pending_gpus:  prometheus.NewDesc("slurm_user_gpus_pending", "Pending gpus for user", labels, nil),
                running: prometheus.NewDesc("slurm_user_jobs_running", "Running jobs for user", labels, nil),
                running_cpus: prometheus.NewDesc("slurm_user_cpus_running", "Running cpus for user", labels, nil),
		running_gpus: prometheus.NewDesc("slurm_user_gpus_running", "Running gpus for user", labels, nil),
                suspended: prometheus.NewDesc("slurm_user_jobs_suspended", "Suspended jobs for user", labels, nil),
        }
}

func (uc *UsersCollector) Describe(ch chan<- *prometheus.Desc) {
        ch <- uc.pending
        ch <- uc.pending_cpus
	ch <- uc.pending_gpus
        ch <- uc.running
        ch <- uc.running_cpus
	ch <- uc.running_gpus
        ch <- uc.suspended
}

func (uc *UsersCollector) Collect(ch chan<- prometheus.Metric) {
        um := ParseUsersMetrics(UsersData())
        for u := range um {
                if um[u].pending > 0 {
                        ch <- prometheus.MustNewConstMetric(uc.pending, prometheus.GaugeValue, um[u].pending, u)
                }
                if um[u].pending_cpus > 0 {
                        ch <- prometheus.MustNewConstMetric(uc.pending_cpus, prometheus.GaugeValue, um[u].pending_cpus, u)
                }
		if um[u].pending_gpus > 0 {
			ch <- prometheus.MustNewConstMetric(uc.pending_gpus, prometheus.GaugeValue, um[u].pending_gpus, u)
		}
                if um[u].running > 0 {
                        ch <- prometheus.MustNewConstMetric(uc.running, prometheus.GaugeValue, um[u].running, u)
                }
                if um[u].running_cpus > 0 {
                        ch <- prometheus.MustNewConstMetric(uc.running_cpus, prometheus.GaugeValue, um[u].running_cpus, u)
                }
		if um[u].running_gpus > 0 {
			//log.Printf("User: %s, Running GPUs: %f", u, um[u].running_gpus) // Ajout du log
			ch <- prometheus.MustNewConstMetric(uc.running_gpus, prometheus.GaugeValue, um[u].running_gpus, u)
		}
                if um[u].suspended > 0 {
                        ch <- prometheus.MustNewConstMetric(uc.suspended, prometheus.GaugeValue, um[u].suspended, u)
                }
        }
}

