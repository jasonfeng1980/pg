package rdb

import (
    "context"
    "github.com/go-redis/redis/v8"
    "github.com/jasonfeng1980/pg/ecode"
    "github.com/jasonfeng1980/pg/util"
    "strings"
    "time"
)

type Key struct{
    CTX  context.Context
    Name string
    Client *RedisConn
    Expr time.Duration
}
type String struct {
    Key
    JoinMode []string
}
type Hash struct {
    Key
    Field string
    JoinMode []string
}
type List struct {
    Key
}
type Set struct {
    Key
}
type ZSet struct {
    Key
}
type GEO struct {
    Key
}
type Stream struct{
    Key
}

// arg 只允许为 map[string]int32|int64|int|float32|float64|float|string
func (h *Hash)Encode(arg map[string]interface{}) (string, error){
    if len(h.JoinMode) == 0 { // 不需要JOIN 就直接返回
        r, err := util.JsonEncode(arg)
        return string(r[:]), err
    }
    var ret  []string
    var newStr string
    for _, k := range h.JoinMode {
        if v, ok := arg[k];ok{
            newStr = util.Str(v)
        } else {
            newStr = ""
        }
        ret = append(ret, strings.ReplaceAll(newStr, "〡", "|" ))
    }
    return strings.Join(ret, "〡"), nil
}

func (h *Hash)Decode(arg string) (ret map[string]string, err error) {
    if len(h.JoinMode) == 0 { // 不需要JOIN 就直接json_decode
        err = util.JsonDecode(arg, ret)
        return
    }
    arr := strings.Split(arg, "〡")
    if len(arr) != len(h.JoinMode) {
        return nil, ecode.RdbWrongDecodeJoin.Error()
    }
    ret = make(map[string]string)
    for k, v := range h.JoinMode {
        ret[v] = arr[k]
    }
    return
}

type hashJoinResult struct {

}

//////////////////////////////////////////////
//
//  Redis-key
//
//////////////////////////////////////////////
// 检查给定 key 是否存在。
func (k *Key) Exists() (int64 ,error){
    return k.Client.W.Exists(k.CTX, k.Name).Result()
}
// 为给定 key 设置过期时间，以秒计。
func (k *Key) Expire() (bool, error){
    return k.Client.W.Expire(k.CTX, k.Name, k.Expr).Result()
}
// ExpireAt 的作用和 EXPIRE 类似，都用于为 key 设置过期时间。 不同在于 ExpireAt 命令接受的时间参数是 UNIX 时间戳(unix timestamp)。
func (k *Key) ExpireAt(tm time.Time) (bool, error){
    return k.Client.W.ExpireAt(k.CTX, k.Name, tm).Result()
}
// 设置 key 的过期时间以毫秒计。
func (k *Key) PExpire() (bool, error){
    return k.Client.W.PExpire(k.CTX, k.Name, k.Expr).Result()
}
// 设置 key 过期时间的时间戳(unix timestamp) 以毫秒计
func (k *Key) PExpireAt(tm time.Time) (bool, error){
    return k.Client.W.PExpireAt(k.CTX, k.Name, tm).Result()
}
// 将当前数据库的 key 移动到给定的数据库 db 当中。
func (k *Key) Move(db int) (bool, error){
    return k.Client.W.Move(k.CTX, k.Name, db).Result()
}
// 该命令用于在 key 存在时删除 key。
func (k *Key) Del() (int64 ,error){
    return k.Client.W.Del(k.CTX, k.Name).Result()
}

//////////////////////////////////////////////
//
//  Redis-string
//
//////////////////////////////////////////////
// 设置指定 key 的值
func (s *String) Set(value interface{}) (string ,error){
    return s.Client.W.Set(s.CTX, s.Name, value, s.Expr).Result()
}
// 只有在 key 不存在时设置 key 的值。
func (s *String) SetNX(value interface{}) (bool, error){
    return s.Client.W.SetNX(s.CTX, s.Name, value, s.Expr).Result()
}
// 对 key 所储存的字符串值，设置或清除指定偏移量上的位(bit)
func (s *String) SetBit(offset int64, value int) (int64 ,error){
    return s.Client.W.SetBit(s.CTX, s.Name, offset, value).Result()
}
// 用 value 参数覆写给定 key 所储存的字符串值，从偏移量 offset 开始。
func (s *String) SetRange(offset int64, value string) (int64 ,error){
    return s.Client.W.SetRange(s.CTX, s.Name, offset, value).Result()
}
// 获取指定 key 的值
func (s *String) Get() (string ,error){
    return s.Client.R.Get(s.CTX, s.Name).Result()
}
// 将给定 key 的值设为 value ，并返回 key 的旧值(old value)。
func (s *String) GetSet(value interface{}) (string ,error){
    return s.Client.W.GetSet(s.CTX, s.Name, value).Result()
}
// 返回 key 中字符串值的子字符
func (s *String) GetRange(start, end int64) (string ,error){
    return s.Client.R.GetRange(s.CTX, s.Name, start, end).Result()
}
// 对 key 所储存的字符串值，获取指定偏移量上的位(bit)。
func (s *String) GetBit(offset int64) (int64 ,error){
    return s.Client.R.GetBit(s.CTX, s.Name, offset).Result()
}
// 返回 key 所储存的字符串值的长度。
func (s *String) StrLen() (int64 ,error){
    return s.Client.R.StrLen(s.CTX, s.Name).Result()
}
// 将 key 中储存的数字值增一。
func (s *String) Incr() (int64 ,error){
    return s.Client.W.Incr(s.CTX, s.Name).Result()
}
// 将 key 所储存的值加上给定的增量值（increment） 。
func (s *String) IncrBy(increment int64) (int64 ,error){
    return s.Client.W.IncrBy(s.CTX, s.Name, increment).Result()
}
// 将 key 所储存的值加上给定的浮点增量值（increment） 。
func (s *String) IncrByFloat(increment float64) (float64 ,error){
    return s.Client.W.IncrByFloat(s.CTX, s.Name, increment).Result()
}
// 将 key 中储存的数字值减一。
func (s *String) Decr() (int64 ,error){
    return s.Client.W.Decr(s.CTX, s.Name).Result()
}
// key 所储存的值减去给定的减量值（decrement） 。
func (s *String) DecrBy(increment int64) (int64 ,error){
    return s.Client.W.DecrBy(s.CTX, s.Name, increment).Result()
}
// 如果 key 已经存在并且是一个字符串， APPEND 命令将指定的 value 追加到该 key 原来值（value）的末尾。
func (s *String) Append(value string) (int64 ,error){
    return s.Client.W.Append(s.CTX, s.Name, value).Result()
}

//////////////////////////////////////////////
//
//  Redis-hash
//
//////////////////////////////////////////////
// 删除一个或多个哈希表字段
func (h *Hash) HDel(field string) (int64 ,error){
    return h.Client.W.HDel(h.CTX, h.Name, field).Result()
}
// 查看哈希表 key 中，指定的字段是否存在。
func (h *Hash) HExists(field string) (bool, error){
    return h.Client.W.HExists(h.CTX, h.Name, field).Result()
}
// 获取存储在哈希表中指定字段的值。
func (h *Hash) HGet(field ...string) (string ,error){
    var (
        f string
    )
    if len(field)==1{
        f = field[0]
    } else {
        f = h.Field
    }
    return h.Client.R.HGet(h.CTX, h.Name, f).Result()
}
// 获取存储在哈希表中指定字段的值。 - 指定Field
func (h *Hash) HGetField() (string ,error){
    return h.Client.R.HGet(h.CTX, h.Name, h.Field).Result()
}
// 获取在哈希表中指定 key 的所有字段和值
func (h *Hash) HGetAll() (map[string]string, error){
    return h.Client.R.HGetAll(h.CTX, h.Name).Result()
}
// 为哈希表 key 中的指定字段的整数值加上增量 increment 。
func (h *Hash) HIncrBy(field string, incr int64) (int64 ,error){
    return h.Client.W.HIncrBy(h.CTX, h.Name, field, incr).Result()
}
// 为哈希表 key 中的指定字段的浮点数值加上增量 increment 。
func (h *Hash) HIncrByFloat(field string, incr float64) (float64 ,error){
    return h.Client.W.HIncrByFloat(h.CTX, h.Name, field, incr).Result()
}
// 获取所有哈希表中的字段。
func (h *Hash) HKeys() ([]string ,error){
    return h.Client.R.HKeys(h.CTX, h.Name).Result()
}
// 获取哈希表中所有值。
func (h *Hash) HVals() ([]string ,error){
    return h.Client.R.HVals(h.CTX, h.Name).Result()
}
// 获取哈希表中字段的数量
func (h *Hash) HLen() (int64 ,error){
    return h.Client.R.HLen(h.CTX, h.Name).Result()
}
// 获取所有给定字段的值
func (h *Hash) HMGet(fields ...string) ([]interface{}, error){
    return h.Client.R.HMGet(h.CTX, h.Name, fields...).Result()
}
// 将哈希表 key 中的 多对 字段 field 的值设为 value
// value 支持一下的形式:
//   - string
//   - map[string]interface{}
func (h *Hash) HSet(value ...interface{}) (int64 ,error){
    var (
        field interface{}
        v interface{}
    )

    if len(value)==1{
        field = h.Field
        v = value[0]
    } else {
        field = value[0]
        v = value[1]
    }
    return h.Client.W.HSet(h.CTX, h.Name, field, v).Result()
}
// 将哈希表 key 中的 多对 字段 field 的值设为 value - 指定filed
func (h *Hash) HSetField(value interface{}) (int64 ,error){
    return h.Client.W.HSet(h.CTX, h.Name, h.Field, value).Result()
}
// 同 HSet
// HMSet 支持一下的形式:
//   - HMSet( "key1", "value1", "key2", "value2")
//   - HMSet([]string{"key1", "value1", "key2", "value2"})
//   - HMSet(map[string]interface{}{"key1": "value1", "key2": "value2"})
func (h *Hash) HMSet(values ...interface{}) (bool, error){
    return h.Client.W.HMSet(h.CTX, h.Name, values...).Result()
}
// 只有在字段 field 不存在时，设置哈希表字段的值。
// 单个添加
func (h *Hash) HSetNX(field string, value interface{}) (bool, error){
    return h.Client.W.HSetNX(h.CTX, h.Name, field, value).Result()
}
// 迭代哈希表中的键值对。
// - cursor - 游标。
// - match - 匹配的模式。
// - count - 指定从数据集里返回多少元素
// 返回的每个元素都是一个元组，每一个元组元素由一个字段(field) 和值（value）组成。
func (h *Hash) HScan(cursor uint64, match string, count int64) ([]string, uint64, error){
    return h.Client.W.HScan(h.CTX, h.Name, cursor, match, count).Result()
}
//////////////////////////////////////////////
//
//  Redis-list
//
//////////////////////////////////////////////
// 移出并获取列表的第一个元素， 如果列表没有元素会阻塞列表直到等待超时或发现可弹出元素为止。
func (l *List) BLPop(timeout time.Duration, keys ...string) ([]string ,error){
    return l.Client.W.BLPop(l.CTX, timeout, keys...).Result()
}
// 移出并获取列表的最后一个元素， 如果列表没有元素会阻塞列表直到等待超时或发现可弹出元素为止。
func (l *List) BRPop(timeout time.Duration, keys ...string) ([]string ,error){
    return l.Client.W.BRPop(l.CTX, timeout, keys...).Result()
}
// 从列表中弹出一个值，将弹出的元素插入到另外一个列表中并返回它； 如果列表没有元素会阻塞列表直到等待超时或发现可弹出元素为止。
func (l *List) BRPopLPush(source, destination string, timeout time.Duration) (string ,error){
    return l.Client.W.BRPopLPush(l.CTX, source, destination, timeout).Result()
}
// 通过索引获取列表中的元素
func (l *List) LIndex(index int64) (string ,error){
    return l.Client.R.LIndex(l.CTX, l.Name, index).Result()
}
// 在列表的元素前或者后插入元素
// - op  = before  ||  after
// - pivot 什么值
func (l *List) LInsert(op string, pivot, value interface{}) (int64 ,error){
    return l.Client.W.LInsert(l.CTX, l.Name, op, pivot, value).Result()
}
// 获取列表长度
func (l *List) LLen() (int64 ,error){
    return l.Client.R.LLen(l.CTX, l.Name).Result()
}
// 移出并获取列表的第一个元素
func (l *List) LPop() (string ,error){
    return l.Client.W.LPop(l.CTX, l.Name).Result()
}
// 将一个或多个值插入到列表头部
func (l *List) LPush(values ...interface{}) (int64 ,error){
    return l.Client.W.LPush(l.CTX, l.Name, values...).Result()
}
// 将一个值插入到【已存在】的列表头部
func (l *List) LPushX(values ...interface{}) (int64 ,error){
    return l.Client.W.LPushX(l.CTX, l.Name, values...).Result()
}
// 获取列表指定范围内的元素
func (l *List) LRange(start, stop int64) ([]string ,error){
    return l.Client.R.LRange(l.CTX, l.Name, start, stop).Result()
}
// 移除列表 count个 等于value 的元素
func (l *List) LRem(count int64, value interface{}) (int64 ,error){
    return l.Client.R.LRem(l.CTX, l.Name, count, value).Result()
}
// 通过索引设置列表元素的值
func (l *List) LSet(index int64, value interface{}) (string ,error){
    return l.Client.W.LSet(l.CTX, l.Name, index, value).Result()
}
// 对一个列表进行修剪(trim)，就是说，让列表只保留指定区间内的元素，不在指定区间之内的元素都将被删除。
func (l *List) LTrim(start, stop int64) (string ,error){
    return l.Client.W.LTrim(l.CTX, l.Name, start, stop).Result()
}
// 移除列表的最后一个元素，返回值为移除的元素。
func (l *List) RPop() (string ,error){
    return l.Client.W.RPop(l.CTX, l.Name).Result()
}
// 移除列表的最后一个元素，并将该元素添加到另一个列表并返回
func (l *List) RPopLPush(source, destination string) (string ,error){
    return l.Client.W.RPopLPush(l.CTX, source, destination).Result()
}
// 在列表中添加一个或多个值
func (l *List) RPush(value ...interface{}) (int64 ,error){
    return l.Client.W.RPush(l.CTX, l.Name, value...).Result()
}
// 为【已存在】的列表添加值
func (l *List) RPushX(value ...interface{}) (int64 ,error){
    return l.Client.W.RPushX(l.CTX, l.Name, value...).Result()
}

//////////////////////////////////////////////
//
//  Redis-set
//
//////////////////////////////////////////////
// 向集合添加一个或多个成员
func (s *Set)SAdd(members ...interface{}) (int64 ,error){
    return s.Client.W.SAdd(s.CTX, s.Name, members...).Result()
}
// 获取集合的成员数
func (s *Set)SCard() (int64 ,error){
    return s.Client.R.SCard(s.CTX, s.Name).Result()
}
// 返回第一个集合与其他集合之间的差异。
func (s *Set)SDiff(keys ...string) ([]string ,error){
    return s.Client.R.SDiff(s.CTX, keys...).Result()
}
// 返回给定所有集合的差集并存储在 destination 中
func (s *Set)SDiffStore(destination string, keys ...string) (int64 ,error){
    return s.Client.W.SDiffStore(s.CTX, destination, keys...).Result()
}
// 返回给定所有集合的交集
func (s *Set)SInter(keys ...string) ([]string ,error){
    return s.Client.W.SInter(s.CTX, keys...).Result()
}
// 返回给定所有集合的交集并存储在 destination 中
func (s *Set)SInterStore(destination string, keys ...string) (int64 ,error){
    return s.Client.W.SInterStore(s.CTX, destination, keys...).Result()
}
// 判断 member 元素是否是集合 key 的成员
func (s *Set)SIsMember(member interface{}) (bool, error){
    return s.Client.R.SIsMember(s.CTX, s.Name, member).Result()
}
// 返回集合中的所有成员
// Redis `SMEMBERS key` command output as a slice.
func (s *Set)SMembers() ([]string ,error){
    return s.Client.R.SMembers(s.CTX, s.Name).Result()
}
// 返回集合中的所有成员
// Redis `SMEMBERS key` command output as a map.
func (s *Set)SMembersMap() (map[string]struct{}, error){
    return s.Client.R.SMembersMap(s.CTX, s.Name).Result()
}
// 将 member 元素从 source 集合移动到 destination 集合
func (s *Set)SMove(destination string, member interface{}) (bool, error){
    return s.Client.W.SMove(s.CTX, s.Name, destination, member).Result()
}
// 移除并返回集合中的1个随机元素
func (s *Set)SPop() (string ,error){
    return s.Client.W.SPop(s.CTX, s.Name).Result()
}
// 移除并返回集合中的多个随机元素
func (s *Set)SPopN(count int64) ([]string ,error){
    return s.Client.W.SPopN(s.CTX, s.Name, count).Result()
}
// 返回集合中一个随机数
func (s *Set)SRandMember() (string ,error){
    return s.Client.R.SRandMember(s.CTX, s.Name).Result()
}
// 返回集合中多个随机数
func (s *Set)SRandMemberN(count int64) ([]string ,error){
    return s.Client.R.SRandMemberN(s.CTX, s.Name, count).Result()
}
// 移除集合中一个或多个成员
func (s *Set)SRem(members ...interface{}) (int64 ,error){
    return s.Client.W.SRem(s.CTX, s.Name, members...).Result()
}
// 返回所有给定集合的并集
func (s *Set)SUnion(keys ...string) ([]string ,error){
    return s.Client.R.SUnion(s.CTX, keys...).Result()
}
// 所有给定集合的并集存储在 destination 集合中
func (s *Set)SUnionStore(keys ...string) (int64 ,error){
    return s.Client.W.SUnionStore(s.CTX, s.Name, keys...).Result()
}
// 迭代集合中的元素
// - cursor - 游标。
// - match - 匹配的模式。
// - count - 指定从数据集里返回多少元素
func (s *Set) SScan(cursor uint64, match string, count int64) ([]string, uint64, error){
    return s.Client.R.SScan(s.CTX, s.Name, cursor, match, count).Result()
}

//////////////////////////////////////////////
//
//  Redis-ZSet (sorted set)
//
//////////////////////////////////////////////
// 向有序集合添加一个或多个成员，或者更新已存在成员的分数
// - nxOrXx
// -    "nx"     => 【只添加新成员。】
// -    "xx"     => 【仅更新存在的成员】
// -    "默认"     => 更新 + 添加
// - returnCh
// -    ture => 修改返回值为发生变化的成员总数
// -    false => 返回新添加成员的总数
// - 更改的元素是新添加的成员，已经存在的成员更新分数。 所以在命令中指定的成员有相同的分数将不被计算在内。
// - 注：在通常情况下，ZADD返回值只计算新添加成员的数量。
func (z *ZSet) ZAdd(nxOrXx string ,returnCh bool, members ...*redis.Z) (int64 ,error){
    if returnCh {
        switch nxOrXx {
        case "nx":
            return z.Client.W.ZAddNXCh(z.CTX, z.Name, members...).Result()
        case "xx":
            return z.Client.W.ZAddXXCh(z.CTX, z.Name, members...).Result()
        default:
            return z.Client.W.ZAddCh(z.CTX, z.Name, members...).Result()
        }
    } else {
        switch nxOrXx {
        case "nx":
            return z.Client.W.ZAddNX(z.CTX, z.Name, members...).Result()
        case "xx":
            return z.Client.W.ZAddXX(z.CTX, z.Name, members...).Result()
        default:
            return z.Client.W.ZAdd(z.CTX, z.Name, members...).Result()
        }
    }
}
// 向有序集合添加一个或多个成员，对有序集合中指定成员的分数加上增量
// - nxOrXx
// -    "nx"     => 【只添加新成员。】
// -    "xx"     => 【仅更新存在的成员】
// -    "默认"     => 更新 + 添加
func (z *ZSet) ZIncr(nxOrXx string , members *redis.Z) (float64 ,error){
    switch nxOrXx {
    case "nx":
        return z.Client.W.ZIncrNX(z.CTX, z.Name, members).Result()
    case "xx":
        return z.Client.W.ZIncrXX(z.CTX, z.Name, members).Result()
    default:
        return z.Client.W.ZIncr(z.CTX, z.Name, members).Result()
    }
}
// 获取有序集合的成员数
func (z *ZSet) ZCard() (int64 ,error){
    return z.Client.R.ZCard(z.CTX, z.Name).Result()
}
// 计算在有序集合中指定区间分数的成员数
func (z *ZSet) ZCount(min, max string) (int64 ,error){
    return z.Client.R.ZCount(z.CTX, z.Name, min, max).Result()
}
// 计算有序集合中指定字典区间内成员数量。
func (z *ZSet) ZLexCount(min, max string) (int64 ,error){
    return z.Client.R.ZLexCount(z.CTX, z.Name, min, max).Result()
}
// 有序集合中对指定成员的分数加上增量 increment
func (z *ZSet) ZIncrBy(increment float64, member string) (float64 ,error){
    return z.Client.W.ZIncrBy(z.CTX, z.Name, increment, member).Result()
}
// 计算给定的一个或多个有序集的交集,并将结果集存储在新的有序集合 destination 中 (member合并，score相加)
func (z *ZSet) ZInterStore(store *redis.ZStore) (int64 ,error){
    return z.Client.W.ZInterStore(z.CTX, z.Name, store).Result()
}
// 删除并返回有序集合key中的最多count个具有【最高得分】的成员。
func (z *ZSet) ZPopMax(count ...int64) ([]redis.Z, error){
    return z.Client.W.ZPopMax(z.CTX, z.Name, count...).Result()
}
// 删除并返回有序集合key中的最多count个具有【最低得分】的成员。
func (z *ZSet) ZPopMin(count ...int64) ([]redis.Z, error){
    return z.Client.W.ZPopMin(z.CTX, z.Name, count...).Result()
}
// 删除并返回非空有序集合 key中分数最大的成员
// BZPOPMAX 是有序集合命令 ZPOPMAX带有阻塞功能的版本。
// 在参数中的所有有序集合均为空的情况下，阻塞连接。参数中包含多个有序集合时，按照参数中key的顺序，返回第一个非空key中分数最大的成员和对应的分数
// 参数 timeout 可以理解为客户端被阻塞的最大秒数值，0 表示永久阻塞。
func (z *ZSet) BZPopMax(timeout time.Duration, keys ...string) (*redis.ZWithKey, error){
    return z.Client.W.BZPopMax(z.CTX, timeout, keys...).Result()
}
// 删除并返回非空有序集合 key中分数最小的成员
func (z *ZSet) BZPopMin(timeout time.Duration, keys ...string) (*redis.ZWithKey, error){
    return z.Client.W.BZPopMin(z.CTX, timeout, keys...).Result()
}
// 通过索引区间返回有序集合指定区间内的成员
func (z *ZSet) ZRange(start, stop int64) ([]string ,error){
    return z.Client.R.ZRange(z.CTX, z.Name, start, stop).Result()
}
// 通过索引区间返回有序集合指定区间内的成员 返回增加 分数
func (z *ZSet) ZRangeWithScores(start, stop int64) ([]redis.Z, error){
    return z.Client.R.ZRangeWithScores(z.CTX, z.Name, start, stop).Result()
}
// 通过分数返回有序集合指定区间内的成员
func (z *ZSet) ZRangeByScore(opt *redis.ZRangeBy) ([]string ,error){
    return z.Client.R.ZRangeByScore(z.CTX, z.Name, opt).Result()
}
// 通过分数返回有序集合指定区间内的成员 返回增加 分数
func (z *ZSet) ZRangeByScoreWithScores(opt *redis.ZRangeBy) ([]redis.Z, error){
    return z.Client.R.ZRangeByScoreWithScores(z.CTX, z.Name, opt).Result()
}
// 通过字典区间返回有序集合的成员
func (z *ZSet) ZRangeByLex(opt *redis.ZRangeBy) ([]string ,error){
    return z.Client.R.ZRangeByLex(z.CTX, z.Name, opt).Result()
}
// 返回有序集合中指定成员的索引
func (z *ZSet) ZRank(member string) (int64 ,error){
    return z.Client.R.ZRank(z.CTX, z.Name, member).Result()
}
// 移除有序集合中的一个或多个成员
func (z *ZSet) ZRem(member string) (int64 ,error){
    return z.Client.W.ZRem(z.CTX, z.Name, member).Result()
}
// 移除有序集合中给定的排名区间的所有成员
func (z *ZSet) ZRemRangeByRank(start, stop int64) (int64 ,error){
    return z.Client.W.ZRemRangeByRank(z.CTX, z.Name, start, stop).Result()
}
// 移除有序集合中给定的分数区间的所有成员
func (z *ZSet) ZRemRangeByScore(min, max string) (int64 ,error){
    return z.Client.W.ZRemRangeByScore(z.CTX, z.Name, min, max).Result()
}
// 移除有序集合中给定的字典区间的所有成员
func (z *ZSet) ZRemRangeByLex(min, max string) (int64 ,error){
    return z.Client.W.ZRemRangeByLex(z.CTX, z.Name, min, max).Result()
}
// 返回有序集中指定区间内的成员，通过索引，分数从高到低
func (z *ZSet) ZRevRange(start, stop int64) ([]string ,error){
    return z.Client.R.ZRevRange(z.CTX, z.Name, start, stop).Result()
}
// 返回有序集中指定分数区间内的成员，分数从高到低排序
func (z *ZSet) ZRevRangeWithScores(start, stop int64) ([]redis.Z, error){
    return z.Client.R.ZRevRangeWithScores(z.CTX, z.Name, start, stop).Result()
}
// 返回有序集中指定分数区间内的成员，分数从高到低排序
func (z *ZSet) ZRevRangeByScore(opt *redis.ZRangeBy) ([]string ,error){
    return z.Client.R.ZRevRangeByScore(z.CTX, z.Name, opt).Result()
}
// 返回有序集中指定分数区间内的成员，分数从高到低排序 返回值 增加 分数
func (z *ZSet) ZRevRangeByScoreWithScores(opt *redis.ZRangeBy) ([]redis.Z, error){
    return z.Client.R.ZRevRangeByScoreWithScores(z.CTX, z.Name, opt).Result()
}
// 返回指定成员区间内的成员，按成员字典倒序排序, 分数必须相同。
func (z *ZSet) ZRevRangeByLex(opt *redis.ZRangeBy) ([]string ,error){
    return z.Client.R.ZRevRangeByLex(z.CTX, z.Name, opt).Result()
}
// 返回有序集合中指定成员的排名，有序集成员按分数值递减(从大到小)排序
func (z *ZSet) ZRevRank(member string) (int64 ,error){
    return z.Client.R.ZRevRank(z.CTX, z.Name, member).Result()
}
// 返回有序集中，成员的分数值
func (z *ZSet) ZScore(member string) (float64 ,error){
    return z.Client.R.ZScore(z.CTX, z.Name, member).Result()
}
// 计算给定的一个或多个有序集的并集，并存储在新的 key 中
func (z *ZSet) ZUnionStore(store *redis.ZStore) (int64 ,error){
    return z.Client.W.ZUnionStore(z.CTX, z.Name, store).Result()
}

// 迭代有序集合中的元素（包括元素成员和元素分值）
// - cursor - 游标。
// - match - 匹配的模式。
// - count - 指定从数据集里返回多少元素
func (z *ZSet) ZScan(cursor uint64, match string, count int64) ([]string, uint64, error){
    return z.Client.R.ZScan(z.CTX, z.Name, cursor, match, count).Result()
}

//////////////////////////////////////////////
//
//  Redis-Geo
//
//////////////////////////////////////////////
// 用于存储指定的地理空间位置，可以将一个或多个经度(longitude)、纬度(latitude)、位置名称(member)添加到指定的 key 中。
func (g *GEO) GeoAdd(geoLocation ...*redis.GeoLocation) (int64 ,error){
    return g.Client.W.GeoAdd(g.CTX, g.Name, geoLocation...).Result()
}
// 用于从给定的 key 里返回所有指定名称(member)的位置（经度和纬度），不存在的返回 nil。
func (g *GEO) GeoPos(members ...string) ([]*redis.GeoPos ,error){
    return g.Client.R.GeoPos(g.CTX, g.Name, members...).Result()
}
// 以给定的经纬度为中心， 返回键包含的位置元素当中， 与中心的距离不超过给定最大距离的所有位置元素
// GeoRadius is a read-only GEORADIUS_RO command.
func (g *GEO) GeoRadius(longitude, latitude float64, query *redis.GeoRadiusQuery) ([]redis.GeoLocation ,error){
    return g.Client.R.GeoRadius(g.CTX, g.Name, longitude, latitude, query).Result()
}
// 根据用户给定的【经纬度坐标】来获取指定范围内的地理位置集合。
// GeoRadiusStore is a writing GEORADIUS command.
func (g *GEO) GeoRadiusStore(longitude, latitude float64, query *redis.GeoRadiusQuery) (int64 ,error){
    return g.Client.R.GeoRadiusStore(g.CTX, g.Name, longitude, latitude, query).Result()
}
// 根据储存在位置集合里面的【某个地点】获取指定范围内的地理位置集合。
// GeoRadius is a read-only GEORADIUSBYMEMBER_RO command.
func (g *GEO) GeoRadiusByMember(member string, query *redis.GeoRadiusQuery) ([]redis.GeoLocation ,error){
    return g.Client.R.GeoRadiusByMember(g.CTX, g.Name, member, query).Result()
}
// 根据储存在位置集合里面的某个地点获取指定范围内的地理位置集合。
// GeoRadiusByMemberStore is a writing GEORADIUSBYMEMBER command.
func (g *GEO) GeoRadiusByMemberStore(member string, query *redis.GeoRadiusQuery) (int64 ,error){
    return g.Client.R.GeoRadiusByMemberStore(g.CTX, g.Name, member, query).Result()
}
// 返回两个给定位置之间的距离。
// - unint:   m | km | mi | ft
func (g *GEO) GeoDist(member1 string, member2 string, unint string) (float64 ,error){
    return g.Client.R.GeoDist(g.CTX, g.Name, member1, member2, unint).Result()
}
// 用于获取多个位置元素的 geohash 值
func (g *GEO) GeoHash(member ...string) ([]string ,error){
    return g.Client.R.GeoHash(g.CTX, g.Name, member...).Result()
}

//////////////////////////////////////////////
//
//  Redis-stream
//
//////////////////////////////////////////////
// 追加消息
func (x *Stream) XAdd(a *redis.XAddArgs) (string ,error){
    return x.Client.W.XAdd(x.CTX, a).Result()
}
// 删除消息，这里的删除仅仅是设置了标志位，不影响消息总长度
func (x *Stream) XDel(ids ...string) (int64 ,error){
    return x.Client.W.XDel(x.CTX, x.Name, ids...).Result()
}
// 消息长度
func (x *Stream) XLen() (int64 ,error){
    return x.Client.W.XLen(x.CTX, x.Name).Result()
}
// 获取消息列表，会自动过滤已经删除的消息 顺序
func (x *Stream) XRange(start, stop string) ([]redis.XMessage ,error){
    return x.Client.W.XRange(x.CTX, x.Name, start, stop).Result()
}
// 获取消息列表，会自动过滤已经删除的消息 顺序【count 数量控制】
func (x *Stream) XRangeN(start, stop string, count int64) ([]redis.XMessage ,error){
    return x.Client.W.XRangeN(x.CTX, x.Name, start, stop, count).Result()
}
// 获取消息列表，会自动过滤已经删除的消息 逆序
func (x *Stream) XRevRange(start, stop string) ([]redis.XMessage ,error){
    return x.Client.W.XRevRange(x.CTX, x.Name, start, stop).Result()
}
// 获取消息列表，会自动过滤已经删除的消息 逆序【count 数量控制】
func (x *Stream) XRevRangeN(start, stop string, count int64) ([]redis.XMessage ,error){
    return x.Client.W.XRevRangeN(x.CTX, x.Name, start, stop, count).Result()
}
// 从一个或者多个流中读取数据，仅返回ID大于调用者报告的最后接收ID的条目。
// - Count 每个流 最多几个元素
// - Block 阻塞调用 超时时间
func (x *Stream) XRead(a *redis.XReadArgs) ([]redis.XStream ,error){
    return x.Client.W.XRead(x.CTX, a).Result()
}
// 从一个或者多个流中读取数据，仅返回ID大于调用者报告的最后接收ID的条目。
// 不堵塞
func (x *Stream) XReadStreams(streams ...string) ([]redis.XStream ,error){
    return x.Client.W.XReadStreams(x.CTX, streams...).Result()
}

// 创建【消费者组】
// - 如果创建组时指定的流不存在，将返回错误
func (x *Stream) XGroupCreate(group, start string) (string ,error){
    return x.Client.W.XGroupCreate(x.CTX, x.Name, group, start).Result()
}
// 创建【消费者组】
// - 如果创建组时指定的流不存在，ID以自动创建流（如果不存在）。
//   请注意，如果以这种方式创建流，则其长度将为0：
func (x *Stream) XGroupCreateMkStream(group, start string) (string ,error){
    return x.Client.W.XGroupCreateMkStream(x.CTX, x.Name, group, start).Result()
}

// 创建【消费者组】 设置下一条要传递的消息
// - XGROUP SETID mystream consumer-group-name 0
func (x *Stream) XGroupSetID(group, start string) (string ,error){
    return x.Client.W.XGroupSetID(x.CTX, x.Name, group, start).Result()
}
// 完全销毁一个【消费者组】
func (x *Stream) XGroupDestroy(group string) (int64 ,error){
    return x.Client.W.XGroupDestroy(x.CTX, x.Name, group).Result()
}
// 将给定的消费者从【消费者组】中删除
func (x *Stream) XGroupDelConsumer(group, consumer string) (int64 ,error){
    return x.Client.W.XGroupDelConsumer(x.CTX, x.Name, group, consumer).Result()
}
// 同XREAD 并支持使用者组
func (x *Stream) XReadGroup(a *redis.XReadGroupArgs) ([]redis.XStream ,error){
    return x.Client.W.XReadGroup(x.CTX, a).Result()
}
// 将指定ID对应的entry从consumer的已处理消息列表中删除
// XACK mystream mygroup 1527864992409-0
func (x *Stream) XAck(group string, ids ...string) (int64 ,error){
    return x.Client.W.XAck(x.CTX, x.Name, group, ids...).Result()
}

func (x *Stream) XPending(group string) (*redis.XPending ,error){
    return x.Client.W.XPending(x.CTX, x.Name, group).Result()
}
func (x *Stream) XPendingExt(a *redis.XPendingExtArgs) ([]redis.XPendingExt ,error){
    return x.Client.W.XPendingExt(x.CTX, a).Result()
}
// 改变待处理消息的所有权
func (x *Stream) XClaim(a *redis.XClaimArgs) ([]redis.XMessage ,error){
    return x.Client.W.XClaim(x.CTX, a).Result()
}
// 改变待处理消息的所有权
// 只返回成功认领的消息ID数组，不返回实际的消息
func (x *Stream) XClaimJustID(a *redis.XClaimArgs) ([]string ,error){
    return x.Client.W.XClaimJustID(x.CTX, a).Result()
}
// 将流裁剪为指定数量的项目  【精准  =maxLen】
func (x *Stream) XTrim(key string, maxLen int64) (int64 ,error){
    return x.Client.W.XTrim(x.CTX, key, maxLen).Result()
}
// 将流裁剪为指定数量的项目  【不精准 >maxLen， 可以多一些 效率更高】
func (x *Stream) XTrimApprox(key string, maxLen int64) (int64 ,error){
    return x.Client.W.XTrimApprox(x.CTX, key, maxLen).Result()
}

func (x *Stream) XInfoGroups(key string) ([]redis.XInfoGroup ,error){
    return x.Client.W.XInfoGroups(x.CTX, key).Result()
}

