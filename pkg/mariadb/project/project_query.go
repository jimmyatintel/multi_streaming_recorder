package project_query

import (
	emailfunction "recorder/internal/email_function"
	"recorder/internal/structure"
	"recorder/pkg/logger"
	dut_query "recorder/pkg/mariadb/dut"
	kvm_query "recorder/pkg/mariadb/kvm"
	"recorder/pkg/mariadb/method"
	"strings"
)

func Update_project_setting(setting structure.Project_setting_Tamplate, Email_string string) {
	_, err := method.Exec("UPDATE project SET short_name = ?, owner=?, email_list=? WHERE project_name = ?", setting.Short_name, setting.Owner, Email_string, setting.Project_name)
	if err != nil {
		logger.Error("Update project status error: " + err.Error())
	}
}
func Get_all_projects_setting() []structure.Project_setting_Tamplate {
	Project, err := method.Query("SELECT project_name,short_name,owner,email_list FROM project")
	if err != nil {
		logger.Error("Query all project error: " + err.Error())
	}
	var project_list []structure.Project_setting_Tamplate
	for Project.Next() {
		var project_setting structure.Project_setting_Tamplate
		var email_string string
		err := Project.Scan(&project_setting.Project_name, &project_setting.Short_name, &project_setting.Owner, &email_string)
		if err != nil {
			logger.Error(err.Error())
			return project_list
		}
		project_setting.Email_list = emailfunction.String_to_Email(email_string)
		project_list = append(project_list, project_setting)
	}
	return project_list
}
func Get_all_Floor() map[string]int {
	Floors := make(map[string]int)
	Projects, err := method.Query("SELECT project_name FROM project")
	if err != nil {
		logger.Error("Query all project error: " + err.Error())
		return Floors
	}
	for Projects.Next() {
		var Floor string
		err = Projects.Scan(&Floor)
		if err != nil {
			logger.Error(err.Error())
			return Floors
		}
		String_list := strings.Split(Floor, "_")
		_, ok := Floors[String_list[0]]
		if ok {
			Floors[String_list[0]]++
		} else {
			Floors[String_list[0]] = 1
		}
	}
	return Floors
}
func Get_duts(project string) []structure.DUT {
	Dut, err := method.Query("SELECT machine_name FROM debug_unit where project=?", project)
	if err != nil {
		logger.Error("Query dut from project error: " + err.Error())
	}
	var DUTS []structure.DUT
	for Dut.Next() {
		var Tmp string
		err = Dut.Scan(&Tmp)
		if err != nil {
			logger.Error(err.Error())
			return DUTS
		}
		d := dut_query.Get_dut_status(Tmp)
		DUTS = append(DUTS, d)
	}
	return DUTS
}
func Get_kvms(project string) []structure.Kvm {
	Kvm, err := method.Query("SELECT hostname FROM debug_unit where project=?", project)
	if err != nil {
		logger.Error("Query dut from project error: " + err.Error())
	}
	var KVMS []structure.Kvm
	for Kvm.Next() {
		var Tmp string
		err = Kvm.Scan(&Tmp)
		if err != nil {
			logger.Error(err.Error())
			return KVMS
		}
		d := kvm_query.Get_kvm_status(Tmp)
		KVMS = append(KVMS, d)
	}
	return KVMS
}
func Get_dbgs(project string) []string {
	Dut, err := method.Query("SELECT ip FROM debug_unit where project=?", project)
	if err != nil {
		logger.Error("Query dut from project error: " + err.Error())
	}
	var IPs []string
	for Dut.Next() {
		var Tmp string
		err = Dut.Scan(&Tmp)
		IPs = append(IPs, Tmp)
	}
	return IPs
}
func Get_Units(project string) []structure.Unit_detail {
	Unit, err := method.Query("SELECT machine_name, hostname, ip FROM debug_unit where project=?", project)
	if err != nil {
		logger.Error("Query dut from project error: " + err.Error())
	}
	var UNITS []structure.Unit_detail
	for Unit.Next() {
		var unit structure.Unit_detail
		var Tmp, Tmp2, Tmp3 string
		err = Unit.Scan(&Tmp, &Tmp2, &Tmp3)
		if err != nil {
			logger.Error(err.Error())
			return UNITS
		}
		unit.Machine_name = dut_query.Get_dut_status(Tmp)
		unit.Ip = Tmp3
		unit.Hostname = kvm_query.Get_kvm_status(Tmp2)
		unit.Project = project
		UNITS = append(UNITS, unit)
	}
	return UNITS
}
