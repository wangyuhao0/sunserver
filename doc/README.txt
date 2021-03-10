#一个以origin引擎创作的游戏服务器 目前包含由引擎创作者提供的登录模块以外，本人追加了房间模块（创建房间，加入房间，退出房间），队列模块（匹配模块），对局模块，可以基本满足正常的逻辑业务模块，底层均以origin本身功能出发，链接:https://github.com/duanhf2012/origin

#使用方式 go run main.go -start nodeid=1 -config ./config/dev/ 启动之后会有cmd，可以addBot 以及 createRoom 等命令可供使用 方便调试

#参数配置 在config/dev/service.json

GateService MsgRouter MsgId 为 路由依据msgId 指向服务 ServiceName 即为指向的服务

QueueService 则为队列模块参数 MaxQueueNum 队列数目 Queue 则为不同的队列 可以决定不同的rank分数匹配 RoomType 房间类型 PlayerNum 为该类型房间参与人数 MinRank MaxRank RankInterval 分别为 最低分 最高分 分数间隔 会依据这3个数值 初始化 队列 例如 0 100 30
则会分为 0-30 30-60 60-90 90-100 100--~ 几个匹配段