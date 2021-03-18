package constpackage

const LoginSql string = "SELECT id, account, nick_name, rank, sex, grade, avatar, coin, create_time, update_time, `password` FROM `user` WHERE account = ?  AND `password` = ?"

const UserRedisKey = "sunserver:userinfo:"

const UserTableName = "user"

const FriendTableName = "friend"
