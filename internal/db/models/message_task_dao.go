package models

import (
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"time"
)

type MessageTaskStatus = int

const (
	MessageTaskStateEnabled  = 1 // 已启用
	MessageTaskStateDisabled = 0 // 已禁用

	MessageTaskStatusNone    MessageTaskStatus = 0 // 普通状态
	MessageTaskStatusSending MessageTaskStatus = 1 // 发送中
	MessageTaskStatusSuccess MessageTaskStatus = 2 // 发送成功
	MessageTaskStatusFailed  MessageTaskStatus = 3 // 发送失败
)

type MessageTaskDAO dbs.DAO

func NewMessageTaskDAO() *MessageTaskDAO {
	return dbs.NewDAO(&MessageTaskDAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "edgeMessageTasks",
			Model:  new(MessageTask),
			PkName: "id",
		},
	}).(*MessageTaskDAO)
}

var SharedMessageTaskDAO *MessageTaskDAO

func init() {
	dbs.OnReady(func() {
		SharedMessageTaskDAO = NewMessageTaskDAO()
	})
}

// 启用条目
func (this *MessageTaskDAO) EnableMessageTask(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", MessageTaskStateEnabled).
		Update()
	return err
}

// 禁用条目
func (this *MessageTaskDAO) DisableMessageTask(tx *dbs.Tx, id int64) error {
	_, err := this.Query(tx).
		Pk(id).
		Set("state", MessageTaskStateDisabled).
		Update()
	return err
}

// 查找启用中的条目
func (this *MessageTaskDAO) FindEnabledMessageTask(tx *dbs.Tx, id int64) (*MessageTask, error) {
	result, err := this.Query(tx).
		Pk(id).
		Attr("state", MessageTaskStateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*MessageTask), err
}

// 创建任务
func (this *MessageTaskDAO) CreateMessageTask(tx *dbs.Tx, recipientId int64, instanceId int64, user string, subject string, body string, isPrimary bool) (int64, error) {
	op := NewMessageTaskOperator()
	op.RecipientId = recipientId
	op.InstanceId = instanceId
	op.User = user
	op.Subject = subject
	op.Body = body
	op.IsPrimary = isPrimary
	op.Status = MessageTaskStatusNone
	op.State = MessageTaskStateEnabled
	return this.SaveInt64(tx, op)
}

// 查找需要发送的任务
func (this *MessageTaskDAO) FindSendingMessageTasks(tx *dbs.Tx, size int64) (result []*MessageTask, err error) {
	if size <= 0 {
		return nil, nil
	}
	_, err = this.Query(tx).
		State(MessageTaskStateEnabled).
		Attr("status", MessageTaskStatusNone).
		Desc("isPrimary").
		AscPk().
		Limit(size).
		Slice(&result).
		FindAll()
	return
}

// 设置发送的状态
func (this *MessageTaskDAO) UpdateMessageTaskStatus(tx *dbs.Tx, taskId int64, status MessageTaskStatus, result []byte) error {
	if taskId <= 0 {
		return errors.New("invalid taskId")
	}
	op := NewMessageTaskOperator()
	op.Id = taskId
	op.Status = status
	op.SentAt = time.Now().Unix()
	if len(result) > 0 {
		op.Result = result
	}
	return this.Save(tx, op)
}
