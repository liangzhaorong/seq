# SPEC: lazy-seq —— Scala 风格惰性迭代器链式操作库

> 技术规格文档，派生自：`tasks/prd-lazy-seq.md`
> 生成日期：2026-06-29 | 目标分支：`main` | Go 版本：`go1.27rc1`（要求 `go 1.27`）

---

## 1. Summary / 概述

### 1.1 本 SPEC 覆盖范围

本文档规定 `github.com/liangzhaorong/seq` 库的**技术实现契约**：基于 Go 1.23 惰性迭代器 `iter.Seq[T]` / `iter.Seq2[K, V]` 的具名包装类型，利用 **Go 1.27 范型方法（type parameters on methods）** 实现 Scala 风格链式集合操作，全面对标 `samber/lo`，并提供并行求值能力。SPEC 定义了类型系统、公共 API 契约、核心算法、错误/并发策略、测试与实现计划——足以让工程师或 AI Agent 据此直接编码，无需再做架构决策。

### 1.2 PRD 参考

- 来源：`tasks/prd-lazy-seq.md`
- 覆盖 User Stories：US-001 ~ US-015（全部）
- 覆盖 Functional Requirements：FR-1 ~ FR-14（全部）

### 1.3 设计决策汇总

| # | 决策项 | 选择 | 理由 |
|---|--------|------|------|
| D1 | Module path | `github.com/liangzhaorong/seq` | 用户确认 |
| D2 | 包结构 | **单包扁平** `package seq` | 链式调用零额外 import；内部实现用 `seq/internal/*` 隔离，用户无感 |
| D3 | 命名风格 | **Scala 风格**（`FlatMap`/`GroupBy`/`ForEach`/`MapValues`/`TakeWhile`） | 用户确认 |
| D4 | **约束操作的组织** | **需约束 `T`（comparable/Ordered/Number）的操作用顶层泛型函数；不约束 `T` 的用链式方法** | Go 1.27 generic method 无法对 receiver 的 `T` 施加额外约束（已验证：`x == v` 报 `incomparable types in type set`）。这是硬性语言限制 |
| D5 | 包装类型 | 具名类型 `type Seq[T any] func(yield func(T) bool)`（非别名） | 类型别名无法定义方法；具名类型底层与 `iter.Seq` 相同，互转零成本 |
| D6 | `Compact` 语义 | 去除零值（`v == *new(T)`），`T comparable` | 对齐 lo `lo.Compact`；去相邻重复用 `Dedupe` |
| D7 | `ToMap` 键冲突 | 默认**后者覆盖**；提供 `ToMapWith(merge)` 变体 | 与 Go map 写入语义一致；merge 满足聚合需求 |
| D8 | `Range` 类型 | v1 仅 `int`（`Range(start, stop, step int) Seq[int]`，stop 不含） | 复杂度可控；数值泛型约束收益有限 |
| D9 | 并行默认 | worker 数 = `GOMAXPROCS(0)`；**默认保序**；`WithOrdered(false)` 关闭 | 默认可预测、可测试；关保序换吞吐 |
| D10 | `SortBy` 返回 | 返回 `Seq[T]`（终止求值时内部物化排序） | 保持链式一致；godoc 标注 O(n) 缓冲代价 |
| D11 | 错误模型 | 不向方法链注入 `error`；编程错误 panic；并发用 ctx 取消 + panic 重抛 | 保持链式 API 纯净 |
| D12 | 运行时依赖 | **零第三方运行时依赖**（仅标准库 + `cmp`） | 库轻量、易审计 |

---

## 2. Architecture / 架构

### 2.1 系统上下文

本库是**纯 Go 库**，无运行时进程、无网络、无存储。它位于用户应用与 Go 标准库 `iter`/`slices`/`maps`/`cmp` 之间，提供一层"可链式、惰性、类型安全"的函数式集合抽象。

```
┌─────────────────────────────────────────────┐
│            用户应用代码                       │
│  seq.Of(...).Map().Filter().Collect()       │
└───────────────────┬─────────────────────────┘
                    │ import "github.com/liangzhaorong/seq"
┌───────────────────▼─────────────────────────┐
│         package seq  (本库，单包)             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ Seq[T]   │  │ Seq2[K,V]│  │ 顶层函数  │   │
│  │ 方法链   │  │ 方法链   │  │ 约束操作  │   │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘   │
│       └─────────────┴────────────┘          │
│              ┌───────────────┐               │
│              │ internal/pool │ (worker pool) │
│              └───────────────┘               │
└───────────────────┬─────────────────────────┘
                    │ 底层 func(yield func(T) bool)
┌───────────────────▼─────────────────────────┐
│      Go 标准库 iter / slices / maps / cmp    │
└─────────────────────────────────────────────┘
```

### 2.2 组件设计

| 组件 | 职责 | 边界 |
|------|------|------|
| **`Seq[T]` / `Seq2[K,V]`** | 核心包装类型；承载所有不约束 `T` 的链式方法 | 不可变、只读流 |
| **构造器（顶层函数）** | 从切片/Map/区间/生成函数/`iter.Seq` 创建 `Seq`/`Seq2` | 唯一的流入口 |
| **链式方法（method）** | 中间操作（惰性）+ 不约束 `T` 的终止操作 | 方法不返回 `error` |
| **顶层泛型函数** | 约束 `T` 的操作（`Contains`/`Uniq`/`Sum`/`Min`...）+ 多流操作（`Zip`/`Union`）+ 工具 | 满足 D4 约束 |
| **`internal/pool`** | worker pool：有界并发、ctx 取消、panic 捕获重抛 | 仅并行操作使用，不对外暴露 |
| **`constraints` 别名** | `Number`/`Ordered`/`Comparable` 约束定义 | 聚合/集合操作的类型约束来源 |

### 2.3 模块交互

**惰性链的求值流程**（以 `Of(1,2,3).Map(f).Filter(p).Collect()` 为例）：

```
Collect() 触发求值
   │ range 最外层 Filter 的 yield 闭包
   ▼
Filter(p):  每收到元素 → 若 p(v) 则 yield(v) 给下游
   │ range Map 的 yield 闭包
   ▼
Map(f):     每收到元素 → yield(f(v)) 给下游
   │ range Of 的底层迭代
   ▼
Of:         从 []T 依次 yield 原始元素
```

每一层是一个闭包，`yield` 返回 `false` 时该层立即 `return`，层层向上传播 break。**无终止操作则全链零执行**。

**并行操作流程**（`ParMap`）：上游 `Seq` → `internal/pool` 分发到 N 个 worker 并发执行 `f` → 结果按配置（保序回填 / 完成即产出）写入输出 `Seq` 的 channel-backed 闭包。

### 2.4 文件结构

```
github.com/liangzhaorong/seq/
├── go.mod                          [NEW] module github.com/liangzhaorong/seq, go 1.27
├── seq.go                          [NEW] Seq[T]/Seq2[K,V] 类型定义 + Unbox 互转
├── constructors.go                 [NEW] Of/FromSlice/FromSeq/Empty/Range/Repeat/Generate/Cycle
├── transform.go                    [NEW] Map/FilterMap/FlatMap/Tap；顶层 Flatten
├── filter.go                       [NEW] Filter/FilterNot；方法 Dedupe/DedupeBy；顶层 Compact
├── slice.go                        [NEW] Take/TakeWhile/Drop/DropWhile/Slice/Head/Tail
├── select.go                       [NEW] First/Last/At（终止）
├── group.go                        [NEW] Chunk/GroupBy/Partition/PartitionBy/CountBy；顶层 Uniq/UniqBy
├── sort.go                         [NEW] SortBy/Reverse/Shuffle/Sample
├── find.go                         [NEW] FindBy/ContainsBy/Any/All/None（方法）；顶层 Find/Contains/IndexOf
├── aggregate.go                    [NEW] Reduce/Fold/Count/ForEach/ForEachWhile/MinBy/MaxBy（方法）
├── collect.go                      [NEW] Collect/ToSlice/Join/String（方法）；顶层 ToMap/ToSet/Sum/Min/Max/Mean
├── zip.go                          [NEW] ZipWith（方法）；顶层 Zip/Unzip/Interleave/CartesianProduct
├── seq2.go                         [NEW] Seq2 构造与方法：Keys/Values/MapKeys/MapValues/MapEntries/FilterKV/ForEachKV
├── set.go                          [NEW] 顶层 Union/Intersect/Difference/ContainsAll
├── parallel.go                     [NEW] 顶层 ParMap/ParFilter/ParForEach/ParReduce + Option
├── constraints.go                  [NEW] Number 约束；Ordered=cmp.Ordered 别名
├── helpers.go                      [NEW] 顶层 Ternary/Coalesce/Retry/RetryWithDelay/Clamp/Identity/DefaultValue
├── internal/
│   └── pool/
│       ├── pool.go                 [NEW] worker pool 实现
│       └── pool_test.go            [NEW]
├── *_test.go                       [NEW] 各文件表驱动单测
├── example_test.go                 [NEW] 每个公开符号的 Example*
├── bench_test.go                   [NEW] 关键操作 benchmark
├── README.md                       [NEW] 特性/安装/快速上手/与 lo 对照表/性能
├── LICENSE                         [NEW]
├── CONTRIBUTING.md                 [NEW]
├── CHANGELOG.md                    [NEW]
└── .github/workflows/ci.yml        [NEW] build/vet/lint/test -race/coverage
```

> 注：单包扁平指**公开 API**全在 `package seq`。`internal/pool` 是内部实现细节，用户无法也不必 import，符合 D2。

---

## 3. Type System / 类型系统

### 3.1 核心类型定义

```go
package seq

// Seq 是 iter.Seq[T] 的具名包装类型，底层为 func(yield func(T) bool)。
// 使用具名类型（非别名）以便定义带类型参数的方法。
type Seq[T any] func(yield func(T) bool)

// Seq2 是 iter.Seq2[K, V] 的具名包装类型。
type Seq2[K any, V any] func(yield func(K, V) bool)
```

### 3.2 与 `iter` 的互转（零成本）

```go
// Unbox 把 Seq[T] 转回 iter.Seq[T]，供标准库 slices/maps 消费。
func (s Seq[T]) Unbox() iter.Seq[T] { return iter.Seq[T](s) }

// FromSeq 把 iter.Seq[T] 包装为 Seq[T]，进入链式世界。
func FromSeq[T any](s iter.Seq[T]) Seq[T] { return Seq[T](s) }
```

底层类型相同，转换为编译期零开销的类型重解释。

### 3.3 类型约束

```go
// Number 约束：所有数值类型（无 complex），供 Sum/Product/Mean 使用。
type Number interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
        ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
        ~float32 | ~float64
}

// Ordered 复用标准库约束。
type Ordered = cmp.Ordered
```

### 3.4 类型关系

- `Seq[T]` ⭄ `iter.Seq[T]`：双向零成本互转。
- `Seq2[K,V]` 的 `Keys()`/`Values()` 产出 `Seq[K]`/`Seq[V]`，衔接到 `Seq` 的全部方法链。
- `Seq[T]` 的 `GroupBy`/`CountBy`/`PartitionBy` 产出 `Seq2[...]`，进入 `Seq2` 链。
- 无继承、无接口嵌套；`Seq`/`Seq2` 是具体类型，方法直接定义其上。

---

## 4. API Design / 公共 API 契约

> **核心规则（D4）**：
> - **方法（method）**：receiver 类型参数 `T any`，**不约束 T**。用于变换/过滤/截取/多数聚合。
> - **顶层函数（func）**：需约束 T（`comparable`/`Ordered`/`Number`），或多流输入。约束操作**无法链式**，以 `seq.Contains(s, v)` 形式调用。

### 4.1 构造器（顶层函数）

| 签名 | 说明 |
|------|------|
| `func Of[T any](vs ...T) Seq[T]` | 字面量构造 |
| `func FromSlice[T any](s []T) Seq[T]` | 切片构造（不拷贝，只读迭代） |
| `func FromSeq[T any](s iter.Seq[T]) Seq[T]` | 标准 `iter.Seq` 包装 |
| `func Empty[T any]() Seq[T]` | 空流 |
| `func Range(start, stop, step int) Seq[int]` | 区间，stop 不含；`step==0` panic |
| `func Repeat[T any](v T, n int) Seq[T]` | 重复元素；`n<0` 视为 0 |
| `func Generate[T any](f func(index int) T, n int) Seq[T]` | 索引生成；`n<0` 视为 0 |
| `func Cycle[T any](s Seq[T], times int) Seq[T]` | 循环 times 次；`times<0` 无限（须配 `Take`） |
| `func FromMap[K comparable, V any](m map[K]V) Seq2[K, V]` | Map → Seq2（顺序非确定，godoc 注明） |

### 4.2 链式方法（中间操作 / 惰性）

| 方法 | 签名 | 备注 |
|------|------|------|
| Map | `Map[U any](f func(T) U) Seq[U]` | 范型方法核心 |
| FilterMap | `FilterMap[U any](f func(T) (U, bool)) Seq[U]` | |
| FlatMap | `FlatMap[U any](f func(T) Seq[U]) Seq[U]` | |
| Tap | `Tap(f func(T)) Seq[T]` | 副作用观察 |
| Filter | `Filter(pred func(T) bool) Seq[T]` | |
| FilterNot | `FilterNot(pred func(T) bool) Seq[T]` | |
| Dedupe | `Dedupe() Seq[T]` | 去相邻重复（需 `T comparable`，故见 D4 注⚠） |
| DedupeBy | `DedupeBy[K comparable](key func(T) K) Seq[T]` | 方法类型参数 K comparable，**可行** |
| Take | `Take(n int) Seq[T]` | |
| TakeWhile | `TakeWhile(pred func(T) bool) Seq[T]` | |
| TakeRight | `TakeRight(n int) Seq[T]` | 需缓冲 n |
| Drop | `Drop(n int) Seq[T]` | |
| DropWhile | `DropWhile(pred func(T) bool) Seq[T]` | |
| DropRight | `DropRight(n int) Seq[T]` | 需缓冲 n |
| Slice | `Slice(start, end int) Seq[T]` | 负索引/越界裁剪，godoc 注明 |
| Chunk | `Chunk(size int) Seq[Seq[T]]` | `size<=0` panic |
| GroupBy | `GroupBy[K comparable](key func(T) K) Seq2[K, Seq[T]]` | |
| PartitionBy | `PartitionBy[K comparable](key func(T) K) Seq2[K, Seq[T]]` | |
| CountBy | `CountBy[K comparable](key func(T) K) Seq2[K, int]` | |
| SortBy | `SortBy(less func(a, b T) bool) Seq[T]` | 终止时物化排序（D10） |
| Reverse | `Reverse() Seq[T]` | 需缓冲 |
| Shuffle | `Shuffle(r *rand.Rand) Seq[T]` | 注入随机源；需缓冲 |
| Sample | `Sample(n int, r *rand.Rand) Seq[T]` | 无放回；需缓冲 |
| ZipWith | `ZipWith[B, U any](b Seq[B], f func(T, B) U) Seq[U]` | 长度不等按短的截断 |

> ⚠ **Dedupe 的特例**：`Dedupe()` 无参、需 `T comparable`。因方法不能约束 receiver 的 T，**`Dedupe` 退化为顶层函数** `func Dedupe[T comparable](s Seq[T]) Seq[T]`；`DedupeBy[K comparable]` 因约束在方法自己的类型参数 K 上，**可保留为方法**。

### 4.3 链式方法（终止操作，不约束 T）

| 方法 | 签名 | 返回 |
|------|------|------|
| Collect | `Collect() []T` | 物化为切片 |
| ToSlice | `ToSlice() []T` | Collect 别名 |
| Join | `Join(sep string, f func(T) string) string` | 用 f 显式格式化（不依赖 Stringer） |
| String | `String() string` | 调试用 `[a b c]` 形式 |
| Count | `Count() int` | |
| ForEach | `ForEach(f func(T))` | |
| ForEachWhile | `ForEachWhile(f func(T) bool)` | f 返回 false 中断 |
| Reduce | `Reduce(merge func(a, b T) T) (T, bool)` | 空流返回零值+false |
| Fold | `Fold[U any](seed U, f func(U, T) U) U` | |
| MinBy | `MinBy(less func(a, b T) bool) (T, bool)` | |
| MaxBy | `MaxBy(less func(a, b T) bool) (T, bool)` | |
| FindBy | `FindBy(pred func(T) bool, def T) T` | 短路 |
| ContainsBy | `ContainsBy(pred func(T) bool) bool` | 短路 |
| Any | `Any(pred func(T) bool) bool` | 短路，存在性 |
| All | `All(pred func(T) bool) bool` | 短路 |
| None | `None(pred func(T) bool) bool` | 短路 |
| First | `First() (T, bool)` | |
| Last | `Last() (T, bool)` | 需迭代到底 |
| At | `At(i int) (T, bool)` | 越界返回零值+false |
| Partition | `Partition(pred func(T) bool) (Seq[T], Seq[T])` | 返回两条惰性流 |

### 4.4 顶层函数（约束 T 或多流）

**集合/约束类：**

| 签名 | 约束 |
|------|------|
| `func Uniq[T comparable](s Seq[T]) Seq[T]` | 全流去重，需缓冲 |
| `func UniqBy[T any, K comparable](s Seq[T], key func(T) K) Seq[T]` | |
| `func Compact[T comparable](s Seq[T]) Seq[T]` | 去零值（D6） |
| `func Contains[T comparable](s Seq[T], v T) bool` | 短路 |
| `func IndexOf[T comparable](s Seq[T], v T) int` | 未找到返回 -1 |
| `func Find[T comparable](s Seq[T], v T, def T) T` | |
| `func ToSet[T comparable](s Seq[T]) map[T]struct{}` | |
| `func ToMap[T any, K comparable, V any](s Seq[T], f func(T) (K, V)) map[K]V` | 键冲突后者覆盖（D7） |
| `func ToMapWith[T any, K comparable, V any](s Seq[T], f func(T) (K, V), merge func(old, new V) V) map[K]V` | 自定义冲突合并 |

**聚合（约束数值/有序）：**

| 签名 | 约束 |
|------|------|
| `func Sum[T Number](s Seq[T]) T` | |
| `func Product[T Number](s Seq[T]) T` | |
| `func Min[T Ordered](s Seq[T]) (T, bool)` | |
| `func Max[T Ordered](s Seq[T]) (T, bool)` | |
| `func Mean[T Number](s Seq[T]) float64` | 空流返回 0 |

**多流组合：**

| 签名 | 说明 |
|------|------|
| `func Zip[A, B any](a Seq[A], b Seq[B]) Seq2[A, B]` | |
| `func Unzip[A, B any](s Seq2[A, B]) (Seq[A], Seq[B])` | |
| `func Flatten[U any](s Seq[Seq[U]]) Seq[U]` | |
| `func Interleave[T any](seqs ...Seq[T]) Seq[T]` | 轮流取 |
| `func CartesianProduct[T any](seqs ...Seq[T]) Seq[[]T]` | 笛卡尔积 |
| `func Union[T comparable](seqs ...Seq[T]) Seq[T]` | 并集 |
| `func Intersect[T comparable](seqs ...Seq[T]) Seq[T]` | 交集 |
| `func Difference[T comparable](a, b Seq[T]) Seq[T]` | 差集 a-b |
| `func ContainsAll[T comparable](a, b Seq[T]) bool` | |

### 4.5 Seq2[K, V] 方法

| 方法 | 签名 |
|------|------|
| Keys | `Keys() Seq[K]` |
| Values | `Values() Seq[V]` |
| MapKeys | `MapKeys[K2 any](f func(K) K2) Seq2[K2, V]` |
| MapValues | `MapValues[V2 any](f func(V) V2) Seq2[K, V2]` |
| MapEntries | `MapEntries[K2, V2 any](f func(K, V) (K2, V2)) Seq2[K2, V2]` |
| FilterKV | `FilterKV(pred func(K, V) bool) Seq2[K, V]` |
| ForEachKV | `ForEachKV(f func(K, V))` |

> Seq2 的 `ToMap`/`Invert` 需 `K`/`V` comparable，按 D4 用顶层函数：`func ToMap[K comparable, V any](s Seq2[K,V]) map[K]V`、`func Invert[K comparable, V comparable](s Seq2[K,V]) Seq2[V,K]`（值冲突后者覆盖，godoc 注明）。

### 4.6 并行操作（顶层函数 + Option）

```go
type Option func(*config)

func WithWorkers(n int) Option        // 默认 GOMAXPROCS(0)；n<=0 回落默认
func WithOrdered(ordered bool) Option // 默认 true（D9）
func WithContext(ctx context.Context) Option
func WithBuffer(n int) Option          // worker 间缓冲，默认 workers 数

func ParMap[T, U any](s Seq[T], f func(T) U, opts ...Option) Seq[U]
func ParFilter[T any](s Seq[T], pred func(T) bool, opts ...Option) Seq[T]
func ParForEach[T any](s Seq[T], f func(T), opts ...Option)
func ParReduce[T any](s Seq[T], merge func(a, b T) T, opts ...Option) (T, bool) // 分片-合并
```

并行操作返回 `Seq`（`ParForEach`/`ParReduce` 除外），可继续链式。

### 4.7 工具函数（顶层，helpers）

| 签名 | 说明 |
|------|------|
| `func Ternary[T any](cond bool, a, b T) T` | |
| `func Coalesce[T comparable](vs ...T) T` | 首个非零值 |
| `func CoalesceOrZero[T any](def T, vs ...T) T` | 全零时返回 def |
| `func Identity[T any](v T) T` | |
| `func DefaultValue[T any]() T` | 零值 |
| `func Clamp[T cmp.Ordered](v, low, high T) T` | |
| `func Retry[T any](attempts int, f func() (T, error)) (T, error)` | attempts<=0 不执行 |
| `func RetryWithDelay[T any](attempts int, delay time.Duration, f func() (T, error)) (T, error)` | |

### 4.8 破坏性变更 / 兼容性

- v0.x：允许破坏性 API 变更（CHANGELOG 记录）。
- v1.0 起：承诺 semver；如需 v2 用 major suffix `/v2`。
- 无现有消费者，首发无破坏性问题。

---

## 5. Business Logic / 核心算法

### 5.1 惰性求值与提前终止（所有中间操作通用）

```text
// 中间操作模板（以 Map 为例）
Map(f):
  返回新闭包 outer(yieldU):
    range 上游 s，对每个 v:
      u := f(v)
      if !yieldU(u):   // 下游不再需要
        return          // 立即停止上游迭代（break 传播）
```

**不变量**：中间操作函数体只在 outer 闭包被 range 时执行；outer 闭包未被 range 时，`f` 调用次数为 0。

### 5.2 数值聚合（Sum）

```text
Sum[T Number](s):
  var acc T = 0
  range s: acc += v
  return acc
```
`Product` 同理（seed=1，`*=`）。`Min`/`Max` 用 `cmp.Compare`，空流返回零值 + `ok=false`。

### 5.3 去重 / 排序 / 集合运算（需缓冲）

- **Uniq**：物化到 `[]T`，用 `map[T]struct{}` 去重保序，再 yield。
- **SortBy**：物化到 `[]T`，`slices.SortFunc`，再 yield（D10）。
- **Union**：所有流物化入一个 set，去重 yield。
- **Intersect**：第一流入 set，后续流各自建 set 取交集。
- **Difference**：`b` 入 set，yield `a` 中不在 set 的元素。

godoc 必须为每个此类操作标注 **"需 O(n) 缓冲，非纯惰性"**。

### 5.4 并行求值（ParMap）

```text
ParMap(s, f, opts):
  cfg := 解析 opts（workers, ordered, ctx, buffer）
  启动 W = workers 个 worker
  启动 1 个 dispatcher：range s，把任务（含原始索引 i）送入 jobCh（有界 buffer）
  每个 worker：从 jobCh 取任务，执行 f(v)（recover 捕获 panic 包装为 panicValue）
      结果（含索引）送入 resCh
  collector goroutine：
    if ordered: 用 []U 按索引回填，顺序 yield（需缓冲全部结果，内存 = n）
    else:       收到即 yield（完成序，内存 ≤ W）
  ctx.Done() → 关闭 jobCh，worker 退出，collector 停止
  任一 worker panic → collector 重抛（用 recover 包装，带原始栈）
  所有 worker 结束 → 关闭 resCh → collector 结束 → 关闭输出 yield
```

**保序内存权衡**（D9）：默认保序需缓冲全部结果，超大流可 `WithOrdered(false)` 换吞吐。

### 5.5 验证规则（编程错误 → panic）

| 条件 | 行为 |
|------|------|
| `Range(step==0)` | panic `"seq: Range step must be non-zero"` |
| `Chunk(size<=0)` | panic `"seq: Chunk size must be positive"` |
| `TakeRight/DropRight(n<0)` | 视为 0 |
| `Repeat/Generate(n<0)` | 视为 0（空流） |
| nil receiver（方法调用） | 允许：等价空流（迭代立即结束），不 panic |
| 并行 worker 内 panic | 捕获并在 collector 重抛 |
| ctx 超时/取消 | 提前结束迭代，已产出结果有效 |

### 5.6 边界用例

- **空流**：所有终止操作返回安全零值（`Reduce`→`(_,false)`，`Min`→`(_,false)`，`Mean`→`0`，`Any`→`false`，`All`→`true`，`Collect`→`nil` 或 `[]T{}`，godoc 注明选 `nil`）。
- **单元素流**：`Reduce`/`Min`/`Max` 返回该元素 + `ok=true`。
- **无限流**（`Cycle(_, -1)`）：仅可配合 `Take`/`Find`/`Any`/`FirstOrDefault` 等短路终止使用；`Collect`/`Sum` 会阻塞/OOM，godoc 警告。
- **键冲突**（`ToMap`）：后者覆盖（D7）。
- **长度不等**（`Zip`/`ZipWith`）：按短的截断。
- **map 顺序**（`FromMap`）：非确定，godoc 注明。

---

## 6. Error Handling / 错误处理

### 6.1 错误分类

| 类别 | 策略 | 示例 |
|------|------|------|
| **编程错误（precondition）** | `panic` 带前缀消息 `seq: ...` | `Range(step=0)`、`Chunk(size<=0)` |
| **空/缺失（normal flow）** | 返回 `(zero, ok bool)` 或安全零值 | `First`/`Find`/`Min`/`Reduce` |
| **并发 panic** | worker 内 `recover` → collector 重抛 | `ParMap` 中 `f` panic |
| **并发取消** | `context.Context` | `WithCanceled` |

### 6.2 重试策略（仅 `Retry` 工具）

`Retry`/`RetryWithDelay` 串行重试 `attempts` 次；非并行、无指数退避（保持简单，godoc 注明可用 `RetryWithDelay` 自行实现退避）。

### 6.3 失败模式

- **goroutine 泄漏**：并行操作必须保证 `ctx` 取消或正常结束时所有 worker/collector 退出（测试用 `runtime.NumGoroutine()` 前后对比断言）。
- **panic 重抛丢失原始栈**：包装时保留 `recover()` 的 value 与调用栈信息。

---

## 7. Concurrency Safety / 并发安全

> 本库为纯库，无认证/授权/数据保护议题。"安全"在此特指**并发正确性**。

### 7.1 数据竞争保证

- **串行操作**：无共享可变状态，天然安全。`Shuffle`/`Sample` 接受外部 `*rand.Rand`，**不**使用全局源，避免并发污染全局锁。
- **并行操作**：worker pool 中 jobCh/resCh/config 为只读共享，结果通过 channel 传递；`go test -race` 必须无竞争（FR-9）。
- **保序回填**：每个结果槽位由唯一 worker 写入，无写冲突。

### 7.2 goroutine 生命周期

- 每个 `ParXxx` 调用启动 `W + 2` 个 goroutine（W worker + 1 dispatcher + 1 collector），随输出流迭代结束而全部退出。
- 不在包级缓存 worker pool（避免跨调用状态泄漏）；如需池化，v1 先用 per-call pool，benchmark 验证后再优化。

---

## 8. Performance / 性能

### 8.1 预期负载

无服务端负载概念。关键维度：**单条流元素数 n**（典型 1k~1M）、**链长度 L**、**并行核数 W**。

### 8.2 优化策略

- **零中间切片**：中间操作不物化，仅 `Sort`/`Uniq`/`Union`/`TakeRight` 等显式缓冲操作例外（godoc 标注）。
- **闭包内联**：保持中间操作闭包短小，利于编译器内联。
- **并行分治**：`ParMap` 用流水线（dispatcher→worker→collector）掩盖延迟。
- **避免反射**：约束操作用泛型 `==`/`+`，不用 `reflect`。

### 8.3 Benchmark 设计（bench_test.go）

| Benchmark | 对比对象 | 目标 |
|-----------|----------|------|
| `Map_Filter_Collect` | `slices` 手写多趟 vs 本库链 | 分配次数 ↓ ≥50%（FR/D8） |
| `Reduce_Sum` | for 循环 vs `seq.Sum` | 耗时接近（< 1.5×） |
| `ParMap_Ordered` | 串行 `Map` | 4 核加速比记录 |
| `ParMap_Unordered` | 串行 `Map` / 保序版 | 加速比 + 内存对比 |
| `SortBy` | `slices.SortFunc` | 耗时接近 |

### 8.4 数据库考量

不适用（无数据库）。

---

## 9. Testing Strategy / 测试策略

### 9.1 单元测试（表驱动）

每个公开符号至少一组表驱动用例，覆盖：正常、边界、空输入、短路。用 `testify`？**否**（零依赖原则，FR-11）——用标准 `testing` + 手写表驱动。

**惰性断言**（关键）：用一个带计数器的 fn 构造流，断言"未调用终止操作时 fn 计数=0"。
```go
calls := 0
s := seq.Generate(func(i int) int { calls++; return i }, 100).Map(func(n int) int { return n })
// 此时断言 calls == 0（惰性）
s.Collect()
// 此时 calls == 100
```

**提前终止断言**：`Take(k)` 后断言上游迭代次数 == k。

### 9.2 集成测试

无外部系统。链式组合测试视为集成：`.Map().Filter().FlatMap().GroupBy().Values().Collect()` 端到端断言结果。

### 9.3 边界与并发测试

- 空流、单元素、无限流 + `Take`、键冲突、长度不等、负索引。
- 并发：`go test -race ./...` 全绿；worker panic 重抛断言；ctx 取消后 goroutine 退出断言（`NumGoroutine` 前后差=0）。

### 9.4 验收标准 → 测试映射

| US/FR | 测试 | 类型 | 描述 |
|-------|------|------|------|
| US-001 / FR-2 | `constructors_test.go` | unit | Of/Range/Repeat 等构造器表驱动 |
| US-002 / FR-3,4 | `transform_test.go` | unit | Map 范型方法 + 惰性断言 |
| US-002 / FR-5 | `laziness_test.go` | unit | 提前终止断言 |
| US-006 / FR-4 | `sort_test.go` | unit | SortBy 物化 + godoc 缓冲标注验证 |
| US-007 / FR-6 | `find_test.go` | unit | Any/All/Contains 短路（断言迭代次数） |
| US-008 / FR-6 | `aggregate_test.go` | unit | Reduce 空流 ok=false |
| US-011 / FR-7 | `seq2_test.go` | unit | Seq2 全方法 + ToMap 键冲突 |
| US-013 / FR-8,9 | `parallel_test.go` | unit+race | ParMap 保序 + `-race` + panic 重抛 |
| US-015 / FR-13 | 全包 `go test -cover` | coverage | ≥ 90% |
| US-015 / FR-12 | `example_test.go` | example | 每公开符号 Example 编译运行 |
| US-015 / FR-14 | `bench_test.go` + CI | bench | bench 可跑 + CI 全绿 |

---

## 10. Implementation Plan / 实现计划

### 10.1 阶段（依赖顺序）

| 阶段 | 内容 | 对应 US | 依赖 |
|------|------|---------|------|
| P0 | 骨架：go.mod、`Seq`/`Seq2` 类型、互转、构造器、`constraints` | US-001 | — |
| P1 | 核心中间操作：Map/FilterMap/FlatMap/Tap/Filter/FilterNot/Take/Drop/Slice | US-002,003,004 | P0 |
| P2 | 终止操作：Collect/Reduce/Fold/Count/ForEach/FindBy/Any/All/First/Last/At | US-007,008,009 | P0 |
| P3 | 顶层约束操作：Contains/Uniq/Compact/Sum/Min/Max/ToMap/ToSet + 多流 Zip/Union | US-009,010,012 | P0 |
| P4 | 分组/排序：Chunk/GroupBy/Partition/CountBy/SortBy/Reverse/Shuffle/Sample | US-005,006 | P1 |
| P5 | Seq2 全方法 + ToMap/Invert | US-011 | P0,P1 |
| P6 | 并行：internal/pool + ParMap/ParFilter/ParForEach/ParReduce + Option | US-013 | P0,P1 |
| P7 | 工具函数：Ternary/Coalesce/Retry/Clamp/Identity | US-014 | P0（独立） |
| P8 | 工程化：补齐 Example、benchmark、CI、README、覆盖率达标 | US-015 | 全部 |

### 10.2 Issue 映射

| Issue | US | SPEC 章节 | 优先级 | 依赖 |
|-------|----|-----------|--------|------|
| #1 | US-001 | §3, §4.1 | 高 | — |
| #2 | US-002 | §4.2, §5.1 | 高 | #1 |
| #3 | US-003 | §4.2 | 高 | #1 |
| #4 | US-004 | §4.2, §4.3 | 高 | #1 |
| #5 | US-007 | §4.3 | 高 | #1 |
| #6 | US-008 | §4.3, §4.4, §5.2 | 高 | #1 |
| #7 | US-009 | §4.3, §4.4, §5.3 | 高 | #1 |
| #8 | US-010 | §4.4 | 中 | #1 |
| #9 | US-011 | §4.5 | 中 | #1 |
| #10 | US-005 | §4.2, §5.3 | 中 | #2 |
| #11 | US-006 | §4.2, §5.3 | 中 | #2 |
| #12 | US-012 | §4.4, §5.3 | 中 | #7 |
| #13 | US-013 | §4.6, §5.4, §7 | 中 | #2 |
| #14 | US-014 | §4.7 | 低 | #1 |
| #15 | US-015 | §8.3, §9, CI | 低 | 全部 |

### 10.3 增量交付

- 每个 Issue 独立 PR，必须：`go build`/`go vet`/`golangci-lint`/`go test` 全绿 + 对应 Example。
- v0.1.0 = P0~P2（核心链可用）即可发布预览版。
- v0.5.0 = 全 US 完成但并行可能待优化。
- v1.0.0 = 覆盖率达标 + benchmark 基线 + API 冻结。

---

## 11. Open Questions & Risks / 待决问题与风险

### 11.1 已决策（待 review 确认，见 §1.3）

D6~D10（`Compact` 语义、`ToMap` 冲突、`Range` 类型、并行默认、`SortBy` 返回）已给出默认选择，review 时可调整。

### 11.2 技术风险

| 风险 | 影响 | 缓解 |
|------|------|------|
| Go 1.27 尚为 rc1，正式版语法可能微调 | 低 | generic method 语法已稳定（已验证 rc1 可用）；锁定 `go 1.27`，CI 用最新 patch |
| 约束操作无法链式（D4）牺牲体验 | 中 | 提供 `ContainsBy`/`UniqBy` 等方法变体（谓词/键形式可不约束 T）覆盖 80% 场景；README 说明设计原因 |
| "纯惰性"宣传与缓冲操作（Sort/Uniq/Union）矛盾 | 中 | godoc 逐操作标注缓冲代价；README 明确"惰性指中间操作，部分操作需物化" |
| 并行默认保序内存高 | 低 | `WithOrdered(false)` 关闭；godoc/README 注明 |
| iter.Seq 与 Seq 互转的认知成本 | 低 | 提供 `FromSeq`/`Unbox` 桥接 + README 图示 |

### 11.3 假设

- 假设 Go 1.27 正式版 generic method 语法与 rc1 完全一致。
- 假设 `GOMAXPROCS(0)` 反映可用 CPU 核数。
- 假设 `Collect` 空流返回 `nil`（而非 `[]T{}`）——lo 返回空切片；本库选 `nil` 以减少分配，**待 review 确认**（影响 `len()==0` 兼容但不影响 `== nil` 检查）。
- 假设单包结构在 API 增长到 ~120 符号后仍可维护（如不可控，v1 前可拆 `seq/par` 子包）。
