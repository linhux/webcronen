# webcronen
------------

A timed task manager based on Go and beego framework development. It is used to uniformly manage the scheduled tasks in the project, providing visual configuration interface, performing log recording, email notification and other functions without relying on the crontab service under *unix.

## Background of the project

This project was developed to solve the problem of a lot of timed tasks in the PHP project of my company, and it is not easy to manage using crontab. The scheduled tasks of my project are also written in PHP and belong to the whole project. I hope that there is a system that can configure these scheduled tasks uniformly, and can view the execution of each task. The task execution completion or failure can automatically remind the development of emails. Personnel, so did this project

## Features

* Unified management of multiple timing tasks.
* Second-level timer, using the time expression of crontab.
* The task can be suspended at any time.
* Record the execution result of each task.
* Execution result email notification.

## Screenshot of the interface

![webcron](https://raw.githubusercontent.com/lisijie/webcron/master/screenshot.png)


## Installation Notes

The system needs to install Go and MySQL.

Get the source code

	$ go get github.com/linhux/webcronen

Open the configuration file conf/app.conf and modify the related configuration.


Create database webcron, then import install.sql

	$ mysql -u username -p -D webcron < install.sql

run

      $ cd go/src/github.com/linhux/webcronen
      $ go build
	$ ./webcronen
	or
	$ nohup ./webcronen 2>&1 > error.log &
	Set to run in the background

access：

http://localhost:8000

user：admin
password：admin888
