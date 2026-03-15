package models

import (
	"testing"
)

func TestTask_TableName(t *testing.T) {
	task := Task{}
	if task.TableName() != "task_task" {
		t.Errorf("期望表名 'task_task', 实际 '%s'", task.TableName())
	}
}

func TestSubtask_TableName(t *testing.T) {
	subtask := Subtask{}
	if subtask.TableName() != "task_subtask" {
		t.Errorf("期望表名 'task_subtask', 实际 '%s'", subtask.TableName())
	}
}

func TestTaskDependency_TableName(t *testing.T) {
	dep := TaskDependency{}
	if dep.TableName() != "task_dependency" {
		t.Errorf("期望表名 'task_dependency', 实际 '%s'", dep.TableName())
	}
}

func TestSubtaskDependency_TableName(t *testing.T) {
	dep := SubtaskDependency{}
	if dep.TableName() != "task_subtask_dependency" {
		t.Errorf("期望表名 'task_subtask_dependency', 实际 '%s'", dep.TableName())
	}
}

func TestRequirement_TableName(t *testing.T) {
	req := Requirement{}
	if req.TableName() != "task_requirement" {
		t.Errorf("期望表名 'task_requirement', 实际 '%s'", req.TableName())
	}
}

func TestRequirementDocument_TableName(t *testing.T) {
	doc := RequirementDocument{}
	if doc.TableName() != "task_requirement_document" {
		t.Errorf("期望表名 'task_requirement_document', 实际 '%s'", doc.TableName())
	}
}

func TestMember_TableName(t *testing.T) {
	member := Member{}
	if member.TableName() != "task_member" {
		t.Errorf("期望表名 'task_member', 实际 '%s'", member.TableName())
	}
}

func TestAssignment_TableName(t *testing.T) {
	assignment := Assignment{}
	if assignment.TableName() != "task_assignment" {
		t.Errorf("期望表名 'task_assignment', 实际 '%s'", assignment.TableName())
	}
}

func TestSubtaskAssignment_TableName(t *testing.T) {
	assignment := SubtaskAssignment{}
	if assignment.TableName() != "task_subtask_assignment" {
		t.Errorf("期望表名 'task_subtask_assignment', 实际 '%s'", assignment.TableName())
	}
}

func TestComment_TableName(t *testing.T) {
	comment := Comment{}
	if comment.TableName() != "task_comment" {
		t.Errorf("期望表名 'task_comment', 实际 '%s'", comment.TableName())
	}
}

func TestActivityLog_TableName(t *testing.T) {
	log := ActivityLog{}
	if log.TableName() != "task_activity_log" {
		t.Errorf("期望表名 'task_activity_log', 实际 '%s'", log.TableName())
	}
}

func TestMessage_TableName(t *testing.T) {
	msg := Message{}
	if msg.TableName() != "task_message" {
		t.Errorf("期望表名 'task_message', 实际 '%s'", msg.TableName())
	}
}

func TestMenu_TableName(t *testing.T) {
	menu := Menu{}
	if menu.TableName() != "task_menu" {
		t.Errorf("期望表名 'task_menu', 实际 '%s'", menu.TableName())
	}
}

func TestConfig_TableName(t *testing.T) {
	cfg := Config{}
	if cfg.TableName() != "task_meta" {
		t.Errorf("期望表名 'task_meta', 实际 '%s'", cfg.TableName())
	}
}

func TestProjectTemplate_TableName(t *testing.T) {
	tpl := ProjectTemplate{}
	if tpl.TableName() != "task_project_template" {
		t.Errorf("期望表名 'task_project_template', 实际 '%s'", tpl.TableName())
	}
}

func TestProjectTemplateTask_TableName(t *testing.T) {
	tpl := ProjectTemplateTask{}
	if tpl.TableName() != "task_project_template_task" {
		t.Errorf("期望表名 'task_project_template_task', 实际 '%s'", tpl.TableName())
	}
}

func TestProjectTemplateSubtask_TableName(t *testing.T) {
	tpl := ProjectTemplateSubtask{}
	if tpl.TableName() != "task_project_template_subtask" {
		t.Errorf("期望表名 'task_project_template_subtask', 实际 '%s'", tpl.TableName())
	}
}

func TestTaskTemplate_TableName(t *testing.T) {
	tpl := TaskTemplate{}
	if tpl.TableName() != "task_template" {
		t.Errorf("期望表名 'task_template', 实际 '%s'", tpl.TableName())
	}
}

func TestBackup_TableName(t *testing.T) {
	backup := Backup{}
	if backup.TableName() != "task_backup" {
		t.Errorf("期望表名 'task_backup', 实际 '%s'", backup.TableName())
	}
}

func TestBackupSchedule_TableName(t *testing.T) {
	schedule := BackupSchedule{}
	if schedule.TableName() != "task_backup_schedule" {
		t.Errorf("期望表名 'task_backup_schedule', 实际 '%s'", schedule.TableName())
	}
}

func TestTaskComplexityReport_TableName(t *testing.T) {
	report := TaskComplexityReport{}
	if report.TableName() != "task_complexity_report" {
		t.Errorf("期望表名 'task_complexity_report', 实际 '%s'", report.TableName())
	}
}

func TestLanguage_TableName(t *testing.T) {
	lang := Language{}
	if lang.TableName() != "task_languages" {
		t.Errorf("期望表名 'task_languages', 实际 '%s'", lang.TableName())
	}
}

func TestJSONMap_Value(t *testing.T) {
	j := JSONMap{"key": "value"}

	val, err := j.Value()
	if err != nil {
		t.Errorf("Value() 返回错误: %v", err)
	}
	if val == nil {
		t.Error("期望 Value() 返回非 nil")
	}
}

func TestJSONMap_Value_Nil(t *testing.T) {
	var j JSONMap
	val, err := j.Value()
	if err != nil {
		t.Errorf("Value() 返回错误: %v", err)
	}
	if val != nil {
		t.Error("期望 Value() 返回 nil")
	}
}

func TestJSONMap_Scan(t *testing.T) {
	var j JSONMap
	err := j.Scan([]byte(`{"key":"value"}`))
	if err != nil {
		t.Errorf("Scan() 返回错误: %v", err)
	}
	if j["key"] != "value" {
		t.Errorf("期望 'value', 实际 '%v'", j["key"])
	}
}

func TestJSONMap_Scan_Nil(t *testing.T) {
	var j JSONMap
	err := j.Scan(nil)
	if err != nil {
		t.Errorf("Scan() 返回错误: %v", err)
	}
	if j != nil {
		t.Error("期望 j 为 nil")
	}
}
