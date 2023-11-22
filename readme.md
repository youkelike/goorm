

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


解析元数据的过程很慢，每次查询都解析一次不合适，考虑把解析结果缓存下来，实现方式包括：
    用一个字典类型的包变量
        扩展性不好，无法给包变量添加方法
        难以测试，除了错误相关，最好不要引入包变量
    用一个结构体（包含一个字典类型的字段）表示元数据注册中心，再把它实例化赋值给一个包变量，作为它的默认实现。类似 http 包的 ServerMux
        扩展性问题解决了，但还是引入了包变量
    用一个结构体 registry（包含一个字典类型的字段）表示元数据注册中心，再抽象出一个 DB 结构体，内嵌注册中心。DB 是连接框架和 sql.DB 包的结合点
注册中心的实现：
    用什么作为字典的 key:
        结构体名，可能出现不同包下同名结构体，但表名不一样
        表名，拿到元数据之前无法知道表名
        reflect.Type，虽然会有同名结构体，但包路径不同，所以合适
注册中心使用普通的字典会有并发问题，解决方法有：
    想办法去掉并发读写的场景，比如 web 框架中注册路由的场景
    使用并发工具，如读写锁 double check，性能稍差
    使用 sync.Map，性能较好，但在 model 不存在时可能会有并发解析和覆盖的问题 

自定义元数据
    自定义数据库表名和列名的实现方式：
        标签，和结构体写一起很内聚，但容易写错，如果是用 protobuf 定义结构体，则要魔改插件 
        接口，接口方法直接加在模型上，但比标签隐晦不好找
        编程注册，用 option 模式
    实现标签定义列名
        指定标签格式，通过反射提取
        通过标签的 Lookup 方法源码可知，它只关心冒号、双引号，冒号前面的是 key，双引号中间的是 value，双引号后面的是另一组 key-value，双引号后面是否有空格无所谓，有空格更易读
    实现接口定义表名
        定义表名获取接口，让模型结构体来实现这个接口
    实现编程方式定义表名、字段名
        抽取一个 Registry 接口（包含 Get 和 Register 两个方法），让注册中心 registry 实现它
        在 register.Register 方法中应用针对 model 的 option 模式，把通过反射解析得到的 model 修改成指定的表名或字段名

跟数据交互的方法 ExecContext、QueryContext、Scan 等都只能接收基本类型的数据，想要自定义一种可用的数据类型（类似 sql.NullString），需要实现两个接口：
    实现了 driver.Valuer 接口的类型可以作为查询参数使用（ExecContext、QueryContext 等的第三个参数）
    实现了 sql.Scanner 接口的类型可以用作 Scan 方法的参数

如何使用事务
如何使用 PrepareStatement
    何时关闭？要有好的使用效果，最好在应用退出时才关闭
    遇到 sql 中的 in 查询，会导致 statement 膨胀
sqlmock 的使用

在框架中整合连接 Database 的功能：
    可以在框架的 DB 结构体中组合 sql.DB
    提供两种初始化数据库连接的方式
        通过传入连接信息来初始化
        通过传入已有连接实例来初始化（主要用于测试）

执行查寻和结果处理
    查询最好统一用 QueryContext，可以让 Get 和 GetMulti 有同样的处理逻辑
    没有查到任何记录时，最好返回错误，和 sql 包保持一致

用反射把查询结果映射到结构体的过程：
    取出查询结果中的所有列名，放到切片中
    遍历列名切片，从元数据中找出列名对应的字段类型信息，创建对应类型的零值变量，放到一个变量切片中
    把变量切片传给 Scan 方法，给每个变量赋值
    通过 new(T) 创建结构体 T 的实例 tp
    再遍历列名切片，从元数据中找出列名对应的字段名，通过反射把变量值赋值给 tp 中的同名字段


unsafe 使用    
    uintptr 类型表示一个内存地址数值，只在计算地址偏移时使用，不要把它赋值给变量，因为 uintptr 对应地址存的数据可能会被 gc 移到其它位置
    unsafe.Pointer 类型不受 gc 的影响（gc 时如果发生了数据移动，会自动更新它指向的地址到新的数据存放位置），可以安全的赋值给变量
    一个对象的内存是连续的，使用对象起始地址+偏移量就可以找到对象中任意字段的地址
    获取对象的起始地址
        reflect.ValueOf(entity).UnsafeAddr() 可以得到对象的起始地址，但垃圾回收后，这个地址存储的数据可能会被移动到其它地方
        address := reflect.ValueOf(entity).UnsafePointer() 不受垃圾回收影响，一般用它
    获取结构体字段的偏移
        fdOffset := reflect.TypeOf(entity).Elem().Field(0).Offset
        fdType   := reflect.TypeOf(entity).Elem().Field(0).Type
    字段值读取
        计算字段的实际地址 fdAddress = unsafe.Pointer(uintptr(address) +  fdOffset)
        如果知道字段实际类型，使用强制转型，比如 *(*int)(fdAddress)
        如果只知道字段的反射类型，用 reflect.NewAt(fdType，fdAddress).Elem().Interface()
    修改字段值
        如果知道字段实际类型，*(*int)(fdAddress) = val
        如果只知道字段的反射类型，reflect.NewAt(fdType，fdAddress).Elem().Set(reflect.ValueOf(val))

用 unsafe 把查询结果映射到结构体的过程
    解析元数据的时候把记录字段的偏移量    
    通过 new(T) 创建结构体 T 的实例 tp，获取它的起始地址
    取出查询结果中的所有列名，放到切片中
    遍历列名切片
        从元数据中找出列名对应的字段类型、偏移量
        用 tp 起始地址+字段偏移，算出字段起始地址
        在字段起始地址位置创建同类型的零值变量，再把变量放到一个变量切片中
    把变量切片传给 Scan 方法，给每个变量赋值，也就相当于给 tp 的每个字段赋值了
    
反射和 unsafe 解析查询结果的时候，步骤差不多，可以重构：
    一种方式是，把解析过程独立做成包方法，反射和 unsafe 各一个包方法，在使用的过程中通过一个 flag 来判断使用谁
        缺点是扩展性不好，比如想给方法加参数会很麻烦
    另一种提取一个 Valuer 抽象，反射和 unsafe 各自提供一个实现，这样更灵活：
        Valuer 中提供两个方法：SetColumns（用于解析查询结果） 和 Field（用于 insert 时根据字段名获取字段值）
        SetColumns 方法中会同时用到返回的查询结果集、模型元数据、new(T)，可以设计成把它们作为参数传入。
        但这样做的话，Field 方法也会要求传入模型元数据作为参数，每次调用会比较繁琐
        一种比较好的方式是，反射和 unsafe 的实现中，都用一个结构体来保存模型元数据和 new(T)，并抽象出一个用于创建 Valuer 接口实现的方法类型 Creator，这个方法就用模型元数据和 new(T) 作为参数，反射和 unsafe 各自都有这个方法类型的创建函数。
        这样的话，使用方式就变成：先调用反射或 unsafe 的创建函数（都是 Creator 类型），获得一个 Valuer 接口的实现，再调用它的 SetColumns 方法进行数据解析
        这样做的一个额外好处是，可以给 DB 加一个 Creator 类型的字段，初始化 DB 的时候就可以指定用哪种方式来解析查询结果（这就是依赖注入？）
由于反射和 unsafe 解析查询结果过程都会用到模型元数据中的字段，所以 Model 和 Field 中的字段都要改成可导出的形式。这样用户就可以提供自己的 Registry 实现了，于是给 DB 增加一个 DBWithRegistry 选项方法，有了这个方法就相当于实现了在多个 DB 中共用 Registry 了（作用类似一个包变量级别的默认 Registry）

在查询中指定列的方法：
    最简设计
        直接传入符合 sql 语法的列名字符串
        传包含多个列名字符串的切片
        缺点包括：都是直接写列名而不是字段名，手误写错了无法校验
    对可以出现在 select 子句后面的对象抽象出一个 Selectable 标记接口
        给 Column 加上 Selectable 接口实现
        实现聚合函数时，也加上 Selectable 接口方法
            把所有聚合函数抽像成一个结构体 Aggregate
            给每个聚合函数定义一个包方法，入参是字段名、返回值都是 Aggregate
        在构造 select 子句时，对 Selectable 接口类型变量进行断言，不同类型 Column、Aggregate 分别处理
如果用户想要在 select 子句中加入子查询或者 'distinct 列名' 之类的写法怎么办？
    不是框架的主流需求可以不支持，但最好有一个简单的兜底方案，比如让用户能直接用原生 sql 语句，但不负责校验
        建立一个 RawExpr 结构体表示所有原生 sql 片段
        让它实现 Selectable 接口就可以用在 select 子句中
        让他实现 Expression 接口、添加加一个 AsPredicate 方法(用于把它转换成 Predicate)，就可以用在 where 子句中作为 Predicate 使用
        Predicate 的 right 字段也是 Expression 类型，所以 RawExpr 也可以作为 Predicate 的一部分使用
            但要加一个 ValueOf 方法，用于在构造 Predicate 时给 right 字段赋值进行类型断言处理
别名情况
    在 Column 和 Aggregate 加入别名支持就好了
        添加 alias 字段
        添加 As 方法
            这个方法做成直接返回一个新的结构体形式，而不是把方法接收器做成指针的形式
                一方面避免修改对象内容后可能引发的内存逃逸
                另一方面返回值是一个新的对象，就像在使用一个不可变对象，避免并发问题
    用户可能会把别名放到 where 子句的 Column 中，不合 sql 语法规范
        方法一可以在构造 where 子句时额外添加一个 flag，通过判断它来决定是否忽略别名
        方法二可以在构造 where 子句时手动把 Column 的别名置空
        原则是，如果你觉得这个地方用户经常会犯错，还难以定位问题，就要在框架中帮他校验，如果某个地方除非是用户存心跟你过不去，不然是不可能用错的，那就不要校验


insert 语句有几种情况要考虑，用逐一追加的方式实现：
    不指定插入列
        要注意插入列的顺序，可以在解析模型元数据的时候把列按解析顺序放到一个切片中，构造 insert 语句时以这个切片的顺序为准
    指定插入列普通列
        把上一步切片中的字段顺序改成指定的字段顺序就行
    指定插入重复键时的更新规则
        这个更新规则的语法和 update 的 set 子句基本一样，可以抽象出一个 Assignable 标记接口
        重构方案有几种：
            方案一，在 Inserter 里增加一个 onConflict 字段用于保存 Assignable 列表，再提供一个方法给它赋值。
                这种方式在 mysql 中很适用，但其它数据库就不行
            方案二，增加一个 OnDuplicateKey 结构体，它包含 Assignable 列表字段，同时又组合到 Inserter 中。再提供一个 builder 模式（OnDuplicateKeyBuilder）来构建包含了 OnDuplicateKey 的 Inserter 实例。这个过程挺绕的，但扩展性更好
        更新规则可以为一个新值，用一个 Assignment 结构体来保存要更新的列和新值，还要让它实现 Assignable 标记接口
        还可以沿用插入的值（列名=values(列名)的形式）。这时候只需让 Column 也实现 Assignable 接口，就可以用到这里。只需在构造 upsert 子句的时候对 Assignable 进行 switch-type，让 Assignment 和 Column 走不同的流程。

方言抽象
    建立一个 Dialect 接口，并提供一个默认实现，不同的数据库继承默认实现，并提供自己的特殊实现
    Dialect 是 DB 级别的抽象，给 DB 增加一个对应字段，并在初始化 DB 的时候注入 Dialect，或者通过 option 模式指定
    增加一个公共的结构体 builder，让它持有 Selector、Insertor、Updator、Deletor 都需要的公共字段，和一些简单的公共方法
    让 Selector、Insertor、Updator、Deletor 都继承 builder，并在创建方法中把 builder 也初始化
    重构相应代码

处理 sqlite 的 upsert 差异
    差异在于它可以指定判断冲突的列
    给 OnDuplicateKey 添加一个字段用于存放冲突列的列表
    给前面 builder 模式的结构体 OnDuplicateKeyBuilder 添加用于指定冲突列的方法 ConflictColumns
    实现 sqlite 版本的 Dialect
重命名
    OnDuplicateKey 改成 Upsert
    OnDuplicateKeyBuilder 改成 UpsertBuilder
Dialect 抽象的缺点
    方言之间不同的地方很多，每个不同的功能点都要求给 Dialect 接口增加方法，接口很容易膨胀。只支持少量通用的方言特性就可以避免
    因为循环引用，Dialect 不能已到 internal 包里
还有一种方案是，提供一个默认的 Insertor，不同方言继承它，并改写各自特有的部分

execute 语句执行结果处理
    可以对 sql.Result 进行包装，建立一个自己的 Result 结构体


事务支持
    建立一个支持事务的结构体 Tx，它实际是 sql.Tx 的代理
        不要把 sql.Tx 嵌入，嵌入可以直接调用 sql.Tx 上的方法，应该让用户调用框架包装后的方法
        用 *sql.Tx 做为字段类型，因为这个结构体比较大，避免复制
    事务是调用 db 上的 beginTx 开启的，因此在 DB 上添加获取 Tx 的方法，注意要让用户传入 context
    NewSelector 这类方法之前只接收 DB 作为参数，要支持现在的事务设计，必须能同时接受 Tx 作为参数
        因为不论事务方式还是非事务方式，sql 包中执行查询方法都是（queryContext、execContext），可以据此抽象出一个接口 Session，可以把它看成 Orm 的会话
        把 NewSelector 参数改成 Session 类型，让 DB 和 Tx 都实现 Session 接口，就都可以传入了
        改造 Selector 时，发现可以提取一个和 DB 共用的结构 core，分别进行改造
    通过闭包的方式管理事务
        在 DB 上添加一个 DoTx 方法，接收用户传入的具体执行函数
        在 DoTx 中开启事务、调用执行函数，并在执行出错、panic 时自动会回滚事务，否则自动提交事务
    在包含事务的相关代码中，经常需要回滚后返回，为防止用户忘记回滚事务就返回了，可以提供一个自动回滚事务的方法 RollbackIfNotCommit
事务扩散方案
    把事务对象放到 context 中传递
    还要给事务对象加一个 done 属性，用于判断事务在传递的过程中是否已经关闭

AOP 方案
    由于框架设计了统一的执行接口 Querier 和 Executor，可以在这两个接口方法的基础上叠加 AOP 逻辑
    抽象出一个 Hanlder，表示要执行的逻辑
    抽象出 QueryContext、QueryResult 作为 Handler 的输入、输出参数
        QueryContext 要组合 QueryBuilder，并有 Type 字段表明是（Selector、Insertor、Updator、Deletor）中的哪一种
        QueryResult 要包含 sql 执行返回结果（sql.Result 或 []*T）和错误信息
    抽象出一个 Middleware，用于串联多个 Handler
    在增删改查的共用结构体上增加字段来保存 Middleware
    给 DB 增加一个选项模式用于添加 Middleware
    改造 Get、GetMulti、Execute 方法，加入对 Middleware 的支持
    实现 log、trace、Prometheus、慢查询、sql 语句审查等 middleware

原生查询支持
    对于框架不支持的 sql 构造特性，提供一种兜底措施，执行查询、结果集的处理还继续依赖框架
    

单元测试目的是确保单一的模块符合预期
集成测试目的是确保模块之间的交互符合预期


支持 join
    建立一个 TableReference 标记接口，作为 join 子句的抽象
    它可以包含三种实现形式：Table、Join、Subquery
    Table 包含模型实例和表别名两个字段
    给 Table 加上 3 种不同的 join 方法（Join、LeftJoin、RightJoin），它们的返回值都是 *JoinBuilder (类似构造 onduplicate 时的模式)，因为一个完整的 join 子句还没结束，需要后续传入 join 的条件，最后应该得到一个 Join 结构体
    Join 本身也可以继续 join，给它也加上这 3 个方法（Join、LeftJoin、RightJoin）
    改造 Selector
        把 table 字段改成 TableReference 类型
        修改 From 方法的参数为 TableReference 类型
        增加一个方法 buildTable 用于构建 from 子句部分，主要是对传入的 TableReference 参数断言（switch-type）处理
        在支持 on 子句时，会出现不同表的不同字段，用原来的 buildColumn 会因为无法切换模型导致找不到字段
            给 Column 结构体加一个字段 table 指向它所在的 Table
            给 Table 结构体加一个方法 C，用于创建一个新的 Column，它包含了字段名和所属表对应的结构体对象
            给 Table 结构体加一个方法 As，用于支持表别名
            改造 buildColumn 方法，通过对传入 Column 对象的 table 字段进行 switch-type 来做不同处理
                断言为 nil 表示 table 字段未赋值，走原来的逻辑
                断言为 Table 表示切换到了新的表，要重新获取模型元数据


Selector 初始化要传入 Session 接口，有两个实现: DB 和 Tx 。
Selector 实现了两个接口： QueryBuilder 和 Querier 。
Selector.Get 方法内部逻辑：
    解析模型元数据，用到 Registry 接口
    调用包方法 get，里面会组装 middleware，把包方法 getHandler 也包装成最底层的 middleware 后调用
    getHandler 方法内部逻辑：
        调用 QueryBuilder.Build 方法获得拼接的 sql 和 args，把它俩传入 Session.queryContext 方法，获得查询结果集 rows
        用 valuer.Creator 和 valuer.Value.SetColumns 把查询结果赋值到对象 

Selector.Build 内部逻辑：
    构造 select 子句
        用标记接口 Selectable 来接收参数，具体可以是 Aggregate、Column、RawExpr，解析的时候断言处理
        在输出列名的时候，要处理列的表名、别名、纯列名 3 种情况
    构造 from 子句
        用标记接口 TableReference 来接收参数，具体可以是 Table、Join、Subquery，解析的时候断言处理
        在输出列名的时候，要处理列的表名、别名、纯列名 3 种情况
    构造 where 子句
        用 Predicate 结构来组织，它的字段用到 Expression 标记接口，具体可以是 Column、Value、Predicate 嵌套、RawExpr，解析的时候断言处理
    构造 group by 子句
        在输出列名的时候，要处理列的表名、别名、纯列名 3 种情况
    构造 having 子句
        同 where 子句的解析一样
    构造 order by 子句
        在输出列名的时候，要处理列的表名、别名、纯列名 3 种情况
    构造 offset limit 子句