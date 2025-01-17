package controllers

import (
	"github.com/astaxie/beego"
	libcron "github.com/linhux/cron"
	"github.com/linhux/webcronen/app/jobs"
	"github.com/linhux/webcronen/app/libs"
	"github.com/linhux/webcronen/app/models"
	"strconv"
	"strings"
	"time"
)

type TaskController struct {
	BaseController
}

// task list
func (this *TaskController) List() {
	page, _ := this.GetInt("page")
	if page < 1 {
		page = 1
	}
	groupId, _ := this.GetInt("groupid")
	filters := make([]interface{}, 0)
	if groupId > 0 {
		filters = append(filters, "group_id", groupId)
	}
	result, count := models.TaskGetList(page, this.pageSize, filters...)

	list := make([]map[string]interface{}, len(result))
	for k, v := range result {
		row := make(map[string]interface{})
		row["id"] = v.Id
		row["name"] = v.TaskName
		row["cron_spec"] = v.CronSpec
		row["status"] = v.Status
		row["description"] = v.Description

		e := jobs.GetEntryById(v.Id)
		if e != nil {
			row["next_time"] = beego.Date(e.Next, "Y-m-d H:i:s")
			row["prev_time"] = "-"
			if e.Prev.Unix() > 0 {
				row["prev_time"] = beego.Date(e.Prev, "Y-m-d H:i:s")
			} else if v.PrevTime > 0 {
				row["prev_time"] = beego.Date(time.Unix(v.PrevTime, 0), "Y-m-d H:i:s")
			}
			row["running"] = 1
		} else {
			row["next_time"] = "-"
			if v.PrevTime > 0 {
				row["prev_time"] = beego.Date(time.Unix(v.PrevTime, 0), "Y-m-d H:i:s")
			} else {
				row["prev_time"] = "-"
			}
			row["running"] = 0
		}
		list[k] = row
	}

	// Group list
	groups, _ := models.TaskGroupGetList(1, 100)

	this.Data["pageTitle"] = "任务列表"
	this.Data["list"] = list
	this.Data["groups"] = groups
	this.Data["groupid"] = groupId
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("TaskController.List", "groupid", groupId), true).ToString()
	this.display()
}

// Add task
func (this *TaskController) Add() {

	if this.isPost() {
		task := new(models.Task)
		task.UserId = this.userId
		task.GroupId, _ = this.GetInt("group_id")
		task.TaskName = strings.TrimSpace(this.GetString("task_name"))
		task.Description = strings.TrimSpace(this.GetString("description"))
		task.Concurrent, _ = this.GetInt("concurrent")
		task.CronSpec = strings.TrimSpace(this.GetString("cron_spec"))
		task.Command = strings.TrimSpace(this.GetString("command"))
		task.Notify, _ = this.GetInt("notify")
		task.Timeout, _ = this.GetInt("timeout")

		notifyEmail := strings.TrimSpace(this.GetString("notify_email"))
		if notifyEmail != "" {
			emailList := make([]string, 0)
			tmp := strings.Split(notifyEmail, "\n")
			for _, v := range tmp {
				v = strings.TrimSpace(v)
				if !libs.IsEmail([]byte(v)) {
					this.ajaxMsg("Invalid email address："+v, MSG_ERR)
				} else {
					emailList = append(emailList, v)
				}
			}
			task.NotifyEmail = strings.Join(emailList, "\n")
		}

		if task.TaskName == "" || task.CronSpec == "" || task.Command == "" {
			this.ajaxMsg("Please complete the information", MSG_ERR)
		}
		if _, err := libcron.Parse(task.CronSpec); err != nil {
			this.ajaxMsg("Invalid cron expression", MSG_ERR)
		}
		if _, err := models.TaskAdd(task); err != nil {
			this.ajaxMsg(err.Error(), MSG_ERR)
		}

		this.ajaxMsg("", MSG_OK)
	}

	// Group list
	groups, _ := models.TaskGroupGetList(1, 100)
	this.Data["groups"] = groups
	this.Data["pageTitle"] = "Add task"
	this.display()
}

// Editing task
func (this *TaskController) Edit() {
	id, _ := this.GetInt("id")

	task, err := models.TaskGetById(id)
	if err != nil {
		this.showMsg(err.Error())
	}

	if this.isPost() {
		task.TaskName = strings.TrimSpace(this.GetString("task_name"))
		task.Description = strings.TrimSpace(this.GetString("description"))
		task.GroupId, _ = this.GetInt("group_id")
		task.Concurrent, _ = this.GetInt("concurrent")
		task.CronSpec = strings.TrimSpace(this.GetString("cron_spec"))
		task.Command = strings.TrimSpace(this.GetString("command"))
		task.Notify, _ = this.GetInt("notify")
		task.Timeout, _ = this.GetInt("timeout")

		notifyEmail := strings.TrimSpace(this.GetString("notify_email"))
		if notifyEmail != "" {
			tmp := strings.Split(notifyEmail, "\n")
			emailList := make([]string, 0, len(tmp))
			for _, v := range tmp {
				v = strings.TrimSpace(v)
				if !libs.IsEmail([]byte(v)) {
					this.ajaxMsg("Invalid email address："+v, MSG_ERR)
				} else {
					emailList = append(emailList, v)
				}
			}
			task.NotifyEmail = strings.Join(emailList, "\n")
		}

		if task.TaskName == "" || task.CronSpec == "" || task.Command == "" {
			this.ajaxMsg("Please complete the information", MSG_ERR)
		}
		if _, err := libcron.Parse(task.CronSpec); err != nil {
			this.ajaxMsg("Invalid cron expression", MSG_ERR)
		}
		if err := task.Update(); err != nil {
			this.ajaxMsg(err.Error(), MSG_ERR)
		}

		this.ajaxMsg("", MSG_OK)
	}

	// Group list
	groups, _ := models.TaskGroupGetList(1, 100)
	this.Data["groups"] = groups
	this.Data["task"] = task
	this.Data["pageTitle"] = "Editing task"
	this.display()
}

// Task execution log list
func (this *TaskController) Logs() {
	taskId, _ := this.GetInt("id")
	page, _ := this.GetInt("page")
	if page < 1 {
		page = 1
	}

	task, err := models.TaskGetById(taskId)
	if err != nil {
		this.showMsg(err.Error())
	}

	result, count := models.TaskLogGetList(page, this.pageSize, "task_id", task.Id)

	list := make([]map[string]interface{}, len(result))
	for k, v := range result {
		row := make(map[string]interface{})
		row["id"] = v.Id
		row["start_time"] = beego.Date(time.Unix(v.CreateTime, 0), "Y-m-d H:i:s")
		row["process_time"] = float64(v.ProcessTime) / 1000
		row["ouput_size"] = libs.SizeFormat(float64(len(v.Output)))
		row["status"] = v.Status
		list[k] = row
	}

	this.Data["pageTitle"] = "Task execution log"
	this.Data["list"] = list
	this.Data["task"] = task
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("TaskController.Logs", "id", taskId), true).ToString()
	this.display()
}

// View log details
func (this *TaskController) ViewLog() {
	id, _ := this.GetInt("id")

	taskLog, err := models.TaskLogGetById(id)
	if err != nil {
		this.showMsg(err.Error())
	}

	task, err := models.TaskGetById(taskLog.TaskId)
	if err != nil {
		this.showMsg(err.Error())
	}

	data := make(map[string]interface{})
	data["id"] = taskLog.Id
	data["output"] = taskLog.Output
	data["error"] = taskLog.Error
	data["start_time"] = beego.Date(time.Unix(taskLog.CreateTime, 0), "Y-m-d H:i:s")
	data["process_time"] = float64(taskLog.ProcessTime) / 1000
	data["ouput_size"] = libs.SizeFormat(float64(len(taskLog.Output)))
	data["status"] = taskLog.Status

	this.Data["task"] = task
	this.Data["data"] = data
	this.Data["pageTitle"] = "View log"
	this.display()
}

// Batch operation log
func (this *TaskController) LogBatch() {
	action := this.GetString("action")
	ids := this.GetStrings("ids")
	if len(ids) < 1 {
		this.ajaxMsg("Please select the item to operate", MSG_ERR)
	}
	for _, v := range ids {
		id, _ := strconv.Atoi(v)
		if id < 1 {
			continue
		}
		switch action {
		case "delete":
			models.TaskLogDelById(id)
		}
	}

	this.ajaxMsg("", MSG_OK)
}

// Batch operation
func (this *TaskController) Batch() {
	action := this.GetString("action")
	ids := this.GetStrings("ids")
	if len(ids) < 1 {
		this.ajaxMsg("Please select the item to operate", MSG_ERR)
	}

	for _, v := range ids {
		id, _ := strconv.Atoi(v)
		if id < 1 {
			continue
		}
		switch action {
		case "active":
			if task, err := models.TaskGetById(id); err == nil {
				job, err := jobs.NewJobFromTask(task)
				if err == nil {
					jobs.AddJob(task.CronSpec, job)
					task.Status = 1
					task.Update()
				}
			}
		case "pause":
			jobs.RemoveJob(id)
			if task, err := models.TaskGetById(id); err == nil {
				task.Status = 0
				task.Update()
			}
		case "delete":
			models.TaskDel(id)
			models.TaskLogDelByTaskId(id)
			jobs.RemoveJob(id)
		}
	}

	this.ajaxMsg("", MSG_OK)
}

// Start task
func (this *TaskController) Start() {
	id, _ := this.GetInt("id")

	task, err := models.TaskGetById(id)
	if err != nil {
		this.showMsg(err.Error())
	}

	job, err := jobs.NewJobFromTask(task)
	if err != nil {
		this.showMsg(err.Error())
	}

	if jobs.AddJob(task.CronSpec, job) {
		task.Status = 1
		task.Update()
	}

	refer := this.Ctx.Request.Referer()
	if refer == "" {
		refer = beego.URLFor("TaskController.List")
	}
	this.redirect(refer)
}

// Suspend the task
func (this *TaskController) Pause() {
	id, _ := this.GetInt("id")

	task, err := models.TaskGetById(id)
	if err != nil {
		this.showMsg(err.Error())
	}

	jobs.RemoveJob(id)
	task.Status = 0
	task.Update()

	refer := this.Ctx.Request.Referer()
	if refer == "" {
		refer = beego.URLFor("TaskController.List")
	}
	this.redirect(refer)
}

// Execute immediately
func (this *TaskController) Run() {
	id, _ := this.GetInt("id")

	task, err := models.TaskGetById(id)
	if err != nil {
		this.showMsg(err.Error())
	}

	job, err := jobs.NewJobFromTask(task)
	if err != nil {
		this.showMsg(err.Error())
	}

	job.Run()

	this.redirect(beego.URLFor("TaskController.ViewLog", "id", job.GetLogId()))
}
