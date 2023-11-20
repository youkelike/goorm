

ORM 框架要解决的问题：
    构造 SQL 语句
    查询结果映射
    事务
    AOP
    关联关系
    方言

核心接口设计风格：
    定义一个大而全的 Orm 接口，各种方法都往里塞。只需要创建一个接口对应的实例，实例本身无状态，可以供给所有模型使用
    定义一个统一的 Query 接口，把构造 SQL 分成多个部分，增删改查方法也都放里面。使用时每个模型各自创建一个接口实例，所以可以在接口上应用泛型
    增删改查分别定义接口，接口中包括中间方法（构造 SQL 某个部分）和终结方法（处理查询结果）
    在上面的基础上优化，不包含中间方法，只有 QueryBuilder(构造 SQL)、 Querier(查) 和 Executor（增删改）三个接口
Select 子句实现：
    这一版暂不考虑，全部用 "SELECT *"
From 子句实现：
    这一版简单处理，把反射解析表名和用户指定表名结合
Where 子句实现：
    常规做法是类似 PrepareStatment 的方式，一个形参是 where 子句字符串（表达式中的参数用问号代替），另一个为变长形参，用于提供 where 子句中的数据，这种做法第二个变长形参容易传错
    好的做法是，对 where 子句中的表达式抽象出一个结构 Predicate
    对于表字段还可以抽象出一个结构 Column    
    构造表达式之间的逻辑关系（and or）这类方法只能定义到 Predicate 上，一元逻辑关系（not）只能定义成函数
    构造表达式内的比较关系（大于小于等于）这类方法最好定义到 Column 上
    进一步，可以把 where 子句部分拆解成一个由 Predicate 作为节点的构成的二叉树
    Predicate 本身由左中右三部分组成，左右两部分既可以是 Column 或具体的值（可以抽象成一个 Value 结构），也可以是嵌套的 Predicate，三者合在一起还可以抽象出一个标记接口 Expression
    
元数据的作用：
    构造 SQL 时字段名校验
    解析查询结果
元数据主要由两个抽象：
    代表表结构的 model
    代表表中列的 field
具体实现：
    测试用例来说明一些不太合理情况，比如表名、字段名的驼峰转换
    限制用户输入来简化代码，比如限制只能传入结构体的一级指针
    出于长远考虑，可以建立一个集中管理 error 的目录
    select 和 delete 在构建 where 子句时，很多代码重复，可以重构，提取一个共用的 builder