package core

import (
	"google.golang.org/protobuf/proto"
	"log"
	"math/rand"
	"mmo_game_zinx/pb"
	"sync"
	"zinx/ziface"
)

// 玩家对象
type Player struct {
	Pid  int32              //玩家ID
	Conn ziface.IConnection //当前玩家的链接（用于和客户端的链接）
	X    float32            //平面x坐标
	Y    float32            //高度
	Z    float32            //平面y坐标(注意不是Y)
	V    float32            //旋转的0-360角度
}

/*
Player ID 生成器
*/
var PidGen int32 = 1  //玩家ID计数器
var IdLock sync.Mutex //保护PidGen的Mutex

// 创建一个玩家的方法
func NewPlayer(conn ziface.IConnection) *Player {
	//生成一个玩家ID
	IdLock.Lock()
	id := PidGen
	PidGen++
	IdLock.Unlock()

	//创建玩家对象
	p := &Player{
		Pid:  id,
		Conn: conn,
		X:    float32(160 + rand.Intn(10)), //随机在160坐标点，基于x轴若干偏移
		Y:    0,
		Z:    float32(140 + rand.Intn(20)), //随机在160坐标点，基于Y轴若干偏移
		V:    0,
	}

	return p
}

/*
提供一向客户端发送消息的方法
将pb的protobuf数据序列化后，再调用zinx的sendMsg
*/
func (p *Player) SendMsg(msgId uint32, data proto.Message) {
	//将proto Message结构体序列化，转化为二进制
	msg, err := proto.Marshal(data)
	if err != nil {
		log.Println("marshal msg err:", err)
		return
	}

	//将msg通过zinx框架sendmsg发给客户端
	if p.Conn == nil {
		log.Println("connection is player is nil")
		return
	}

	if err = p.Conn.SendMsg(msgId, msg); err != nil {
		log.Println("Player SendMsg error!")
	}
}

// 告知客户端玩家Pid
func (p *Player) SyncPid() {
	//组建MsgID:0 的proto数据
	proto_data := &pb.SyncPid{
		Pid: p.Pid,
	}

	//将消息发送客户端
	p.SendMsg(1, proto_data)
}

// 广播玩家字节的初始位置
func (p *Player) BroadCastStartPosition() {
	//组件MsgID:200 的proto数据
	proto_data := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  2, //tp:2代表广播位置坐标
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}

	//将消息发送客户端
	p.SendMsg(200, proto_data)
}

// 玩家广播世界聊天
func (p *Player) Talk(context string) {
	//组件msgID:200 proto数据
	proto_msg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  1,
		Data: &pb.BroadCast_Content{
			Content: context,
		},
	}
	//得到当前世界所有的在线玩家
	players := WorldMgrObj.GetAllPlayers()

	//向所有玩家发送聊天数据
	for _, player := range players {
		//分别给对应的客户端发送消息
		player.SendMsg(200, proto_msg)
	}
}
