# ORM
一个有完整功能的 orm 框架

# 特性
    支持模型元数据解析
    支持结构化构造基本查询、JOIN 查询，支持原生 sql 和以原生 sql 片段的方式构造的子查询
    支持在构建查询的过程中对各个位置的字段名进行校验
    支持 upsert 方言
    支持通过 reflect 和 unsafe 两种方式进行结果集映射
    支持事务
    支持 AOP

# 元数据解析
    通过 reflect 解析模型元数据，用元数据注册中心缓存解析结果
    支持通过标签定义列名、通过接口定义表名、通过选项模式修改表名、字段名

# JOIN 支持
    通过建立一个 TableReference 标记接口作为 join 子句的抽象，用 builder 模式递归构造
    Selector.Scan 方法专用于解析涉及多个模型的 join 查询结果

# 结果集映射
    以接口形式支持结果集映射，可以在 reflect 和 unsafe 两种方案中切换

# 事务支持
    可以手动开启事务，也可以通过闭包的方式使用事务，还提供了自动回滚事务的方法防止用户忘记回滚事务就返回了

# AOP 支持
    通过 AOP 实现对 log、trace、Prometheus、慢查询、sql 语句审查等中间件的支持

# 使用示例
### 获取 db 对象
```go
db, err := Open("sqlite3", "file:test.db?cache=shared&mode=memory", opts...)
```

### 查询
```go
NewSelector[TestModel](db).Where(C("Id").Eq(1)).Get()
NewSelector[TestModel](db).Where(C("Age").Eq(18).And(C("FirstName").Eq("Tom"))).Get()
NewSelector[TestModel](db).Where(C("Age").Eq(18).Or(C("FirstName").Eq("Tom"))).Get()
NewSelector[TestModel](db).
    Where(C("Age").Eq(18).Or(C("FirstName").Eq("Tom"))).
    GroupBy(C("Age")).
    OrderBy(C("Age").Asc()).
    Offset(10).
    Limit(10).
    GetMulti()

使用聚合函数
NewSelector[TestModel](db).Select(Sum(C("Age")), Count(C("FirstName"))).Get()
NewSelector[TestModel](db).Select(Sum(TableOf(new(TestModel)).As("t").C("Age"))).Get()

使用原生 sql 片段
NewSelector[TestModel](db).Select(Raw("COUNT(DISTINCT first_name)")).Get()
NewSelector[TestModel](db).Where(Raw("age>?", 18).AsPredicate()).Get()
NewSelector[TestModel](db).Where(C("Id").Eq(Raw("age+?", 1))).Get()

JOIN 查询
t1 := TableOf(&Order{}).As("t1")
t2 := TableOf(&OrderDetail{}).As("t2")
t3 := t1.RightJoin(t2).On(t1.C("Id").Eq(t2.C("OrderId")))
NewSelector[Order](db).
  Select(t1.C("Id"), t2.C("ItemId"), t2.C("Price"), t2.C("Address")).
  From(t3).
  Where(t1.C("Id").Gt(100)).
  Scan(&Result{})

t1 := TableOf(&Order{}).As("t1")
t2 := TableOf(&OrderDetail{}).As("t2")
t3 := t1.RightJoin(t2).On(t1.C("Id").Eq(t2.C("OrderId")))
NewSelector[Result](db).
    Select(t1.C("Id"), t2.C("ItemId"), t2.C("Price"), t2.C("Address")).
    From(t3).
    Where(t1.C("Id").Gt(100)).
    GetMulti(context.Background())
```
### 插入
```go
简单插入记录
NewInserter[TestModel](db).Values(&TestModel{}).Exec()
NewInserter[TestModel](db).Columns("Id", "FirstName").Values(&TestModel{
	Id:        1,
	FirstName: "Tom",
}, &TestModel{
	Id:        2,
	FirstName: "Tom2",
}).Exec()

使用 upsert
NewInserter[TestModel](db).
    Values(&TestModel{
        Id:        1,
        FirstName: "Tom",
        Age:       18,
        LastName:  &sql.NullString{Valid: true, String: "Jerry"},
    }).
    Upsert().
    Update(Assign("Age", 10), Assign("FirstName", "Bob")).
    Exec()
```
### 更新
```go
NewUpdater[TestModel](db).Value(tm)
指定更新条件
NewUpdater[TestModel](db).
    Value(tm).
    Where(C("FirstName").Eq("Tom")).
    Exec()
指定更新列
NewUpdater[TestModel](db).
    Value(tm).
    Updates(C("Age"), C("FirstName")).
    Where(C("FirstName").Eq("Tom")).
    Exec()
```
### 删除
```go

NewDeletor[TestModel](db).Where(C("FirstName").Eq("Tom")).Exec()
NewDeletor[TestModel](db).
    Where(C("FirstName").Eq("Tom").And(C("Age").Eq(18))).
    Exec()
NewDeletor[TestModel](db).
    From("test_db.test_model").
    Where(C("FirstName").Eq("Tom").And(C("Age").Eq(18))).
    Exec()
```
### 原生查询
```go
RawQuery[TestModel](db, "SELECT * FROM test_model WHERE id = ?", -1).Get()

```