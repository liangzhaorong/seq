# PRD: lazy-seq —— Scala 风格的惰性迭代器链式操作库

## 1. Introduction / 概述

本库为 Go 语言提供一个**面向 `iter.Seq[T]` / `iter.Seq2[K, V]` 的、Scala 风格的可链式调用集合操作库**，定位对标 [`samber/lo`](https://github.com/samber/lo)。

它依赖两项现代 Go 能力：

- **惰性迭代器（Go 1.23+）**：以标准库 `iter.Seq[T]`（即 `func(yield func(T) bool)`）为底层抽象，所有中间操作都是惰性的，仅在终止操作时求值，天然支持"按需拉取 / 提前 break"。
- **范型方法（Go 1.27）**：利用 Go 1.27 新增的"方法上的类型参数（type parameters on methods）"特性，让方法能够拥有独立的类型参数，从而实现真正的链式类型变换（例如 `Map[U]` 返回 `Seq[U]`）。这是过去 Go 无法做到的，也是本库区别于 lo 的核心创新点。

最终用户可以写出如下的 Scala 风格代码：

```go
seq.Of(1, 2, 3, 4, 5).
    Filter(func(n int) bool { return n%2 == 1 }).
    Map(func(n int) string { return fmt.Sprintf("#%d", n) }).
    Take(2).
    Collect() // → ["#1", "#3"]
```

## 2. Goals / 目标

- 提供一个基于 `iter.Seq[T]` 的包装类型 `Seq[T]`，并通过 Go 1.27 范型方法实现**完整的链式调用**能力。
- **全面对标 samber/lo**，覆盖其主要的集合操作类别（映射、过滤、截取、分组、排序、查找、聚合、Zip、Seq2、集合运算、工具函数等），并提供丰富的**终止操作**完成物化（`Collect`/`ToMap`/`Reduce` 等）。
- 提供一组**并行操作**（`ParMap` / `ParFilter` 等），基于 worker pool 实现并发求值。
- 保持**纯惰性、零中间切片**的中间操作语义，避免不必要的内存分配。
- 作为**正式开源库**交付：完整表驱动单测、example、benchmark、CI、lint、godoc 文档，并承诺 API 兼容性（semver）。
- 对外暴露的 API 仅依赖 Go 标准库，**零第三方运行时依赖**。

## 3. User Stories / 用户故事

> 约定：本库为无 UI 的 Go 库，所有故事的验收均以 `go build` / `go vet` / `golangci-lint` / 单测 / benchmark / example 编译通过为准。

### US-001: 项目骨架与包装类型 + 构造器
**Description:** 作为库的使用者，我希望有一组构造器把任意数据源（字面量、切片、Map、range 区间、生成函数）转换成可链式调用的 `Seq[T]` / `Seq2[K, V]`，以便后续链式操作。

**Acceptance Criteria:**
- [ ] 初始化 `go.mod`（Go 版本声明为 `go 1.27`），确认 module path（见 Open Questions）
- [ ] 定义具名包装类型 `type Seq[T any] func(yield func(T) bool)` 与 `type Seq2[K any, V any] func(yield func(K, V) bool)`，可安全与 `iter.Seq` 互转
- [ ] 实现构造器：`Of[T](...T)`、`FromSeq[T](iter.Seq[T])`、`FromSlice[T]([]T)`、`Empty[T]()`
- [ ] 实现区间/生成构造器：`Range(start, stop, step)`、`Repeat[T](v, n)`、`Generate[T](func(i int) T, n)`、`Cycle[T](Seq[T])`（有限次循环可控）
- [ ] 提供 `Unbox()` / 隐式转换，使 `Seq[T]` 可被标准库 `slices`/`maps` 消费
- [ ] `go build ./...`、`go vet ./...`、`golangci-lint run` 全部通过
- [ ] 表驱动单测覆盖：空输入、单元素、负 step、n=0、生成函数边界

### US-002: 映射与扁平化（中间操作）
**Description:** 作为使用者，我希望对流的每个元素做变换或展开，以便在不分配中间切片的情况下做数据塑形。

**Acceptance Criteria:**
- [ ] 实现 `Map[U](func(T) U) Seq[U]`（验证 Go 1.27 范型方法可编译运行）
- [ ] 实现 `FilterMap[U](func(T) (U, bool)) Seq[U]`
- [ ] 实现 `FlatMap[U](func(T) Seq[U]) Seq[U]` 与 `Flatten[U](Seq[Seq[U]])` 构造器辅助
- [ ] 实现 `Tap(func(T)) Seq[T]`（副作用观察，不影响流）
- [ ] 所有中间操作验证惰性：未调用终止操作时，变换函数 0 次执行
- [ ] 中间操作验证提前终止：`Take(n)` 后下游 yield 返回 false 时上游不再迭代
- [ ] 通过 `go vet` / `golangci-lint`；表驱动单测覆盖空流、单元素、嵌套流、提前 break

### US-003: 过滤与压缩（中间操作）
**Description:** 作为使用者，我希望按谓词筛选或压缩流中的无效值。

**Acceptance Criteria:**
- [ ] 实现 `Filter(func(T) bool)`、`FilterNot(func(T) bool)`
- [ ] 实现 `Compact[T comparable]()`（去除连续/全部重复，参照 lo 语义并在 godoc 注明）
- [ ] 实现 `Dedupe()` / `DedupeBy(func(T) K)`（去除相邻重复 / 按键去重的相邻重复）
- [ ] 通过 lint；单测覆盖全保留、全过滤、空流、可比与不可比类型

### US-004: 截取与选取（中间操作）
**Description:** 作为使用者，我希望取头部、尾部或按条件截取流的子区间。

**Acceptance Criteria:**
- [ ] 实现 `Take(n)`、`TakeWhile(func(T) bool)`、`TakeRight(n)`
- [ ] 实现 `Drop(n)`、`DropWhile(func(T) bool)`、`DropRight(n)`
- [ ] 实现 `First() (T, bool)`、`Last() (T, bool)`、`At(i) (T, bool)`、`Head()`、`Tail()`
- [ ] 实现 `Slice(start, end)`（子区间，支持负索引或越界裁剪，godoc 注明语义）
- [ ] 单测覆盖 n=0、n>len、负数、空流；通过 lint

### US-005: 去重、分组、分块（中间操作 + Seq2 衔接）
**Description:** 作为使用者，我希望对流去重、按键分组、分块、分区。

**Acceptance Criteria:**
- [ ] 实现 `Uniq[T comparable]()`、`UniqBy[K](func(T) K)`
- [ ] 实现 `GroupBy[K](func(T) K) Seq2[K, Seq[T]]`（分组结果流入 Seq2，衔接 US-011）
- [ ] 实现 `Chunk(size) Seq[Seq[T]]`
- [ ] 实现 `Partition(func(T) bool) (Seq[T], Seq[T])` 与 `PartitionBy[K](func(T) K) Seq2[K, Seq[T]]`
- [ ] 实现 `CountBy[K](func(T) K) Seq2[K, int]`
- [ ] 单测覆盖 size=1、size>len、单分组、空流；通过 lint

### US-006: 排序与乱序（中间操作）
**Description:** 作为使用者，我希望对元素排序、反转或随机打乱。

**Acceptance Criteria:**
- [ ] 实现 `SortBy(less func(a, b T) bool)`（注：排序需缓冲，godoc 注明其需物化，仍返回 `Seq[T]`）
- [ ] 实现 `OrderBy`（多 key 排序辅助）或 `Sorted(cmp)` 等价物
- [ ] 实现 `Reverse()`
- [ ] 实现 `Shuffle(r *rand.Rand)`（注入随机源以保证可测试）
- [ ] 实现 `Sample(n)`（无放回抽样）
- [ ] 单测覆盖稳定排序、逆序、单元素、空流、Shuffle 可复现（固定 seed）；通过 lint

### US-007: 查找与包含（终止操作）
**Description:** 作为使用者，我希望在流中查找元素或判断存在性。

**Acceptance Criteria:**
- [ ] 实现 `Find(default T) T`、`FindBy(func(T) bool, default T) T`、`FindIndexOf` 等价物
- [ ] 实现 `Contains(v T) bool`（`T comparable`）、`ContainsBy(func(T) bool) bool`
- [ ] 实现 `IndexOf(v T) int`、`LastIndexOf`
- [ ] 实现 `Any(func(T) bool) bool`（存在性）、`All(func(T) bool) bool`、`None(...)`
- [ ] 实现短路语义：`Any`/`All`/`Contains` 命中后立即停止迭代（单测断言迭代次数）
- [ ] 单测覆盖命中、未命中、空流、默认值返回；通过 lint

### US-008: 聚合与统计（终止操作）
**Description:** 作为使用者，我希望对流做归约与统计，得到标量结果。

**Acceptance Criteria:**
- [ ] 实现 `Reduce(merge func(a, b T) T) (T, bool)`、`Fold[U](seed U, func(U, T) U) U`、`ReduceRight`
- [ ] 实现 `Count() int`、`ForEach(func(T))`、`ForEachWhile(func(T) bool)`（可中断）
- [ ] 实现 `Sum()` / `Product()` / `Min()` / `Max()`（数值约束）、`MinBy` / `MaxBy`（自定义比较）
- [ ] 实现 `Mean()`（必要时返回浮点，godoc 注明）
- [ ] 单测覆盖空流归约的 `ok=false`、单元素、浮点累加；通过 lint

### US-009: 物化（终止操作）
**Description:** 作为使用者，我希望把惰性流落盘为具体集合类型。

**Acceptance Criteria:**
- [ ] 实现 `Collect() []T`（等价 `ToSlice`）、`ToSlice()`
- [ ] 实现 `ToMap[K, V](func(T) (K, V)) map[K] V`、`ToSet[T comparable]() map[T] struct{}`
- [ ] 实现 `Join(sep string) string`（`T` 需为 string 或提供 `Stringer`/格式化函数重载，godoc 注明）
- [ ] 实现 `String()`（`fmt.Stringer`，用于调试输出）
- [ ] 单测覆盖大流、空流、键冲突（`ToMap` 覆盖策略注明）；通过 lint

### US-010: Zip / Interleave / 笛卡尔积（中间操作）
**Description:** 作为使用者，我希望把多个流按位置组合或交错。

**Acceptance Criteria:**
- [ ] 实现顶层 `Zip[A, B](Seq[A], Seq[B]) Seq2[A, B]` 与 `Seq.ZipWith[B, U](Seq[B], func(A, B) U) Seq[U]`
- [ ] 实现 `Unzip[A, B](Seq2[A, B]) (Seq[A], Seq[B])`
- [ ] 实现 `Interleave(seqs ...Seq[T]) Seq[T]`（轮流取元素）
- [ ] 实现 `CartesianProduct(seqs ...Seq[T]) Seq[[]T]`
- [ ] 单测覆盖长度不等（截断策略注明）、空流、单流；通过 lint

### US-011: Seq2[K, V] 操作（KV 流）
**Description:** 作为使用者，我希望对 `Seq2[K, V]`（即 `iter.Seq2`）做与 Seq 对等的链式操作。

**Acceptance Criteria:**
- [ ] 实现 `Seq2[K, V]` 包装类型与构造器 `FromMap(map[K]V)`、`FromSeq2(iter.Seq2[K, V])`
- [ ] 实现 `Keys() Seq[K]`、`Values() Seq[V]`、`MapKeys(func(K) K2)`、`MapValues(func(V) V2)`、`MapEntries(func(K, V) (K2, V2))`
- [ ] 实现 `FilterKV(func(K, V) bool)`、`Invert() Seq2[V, K]`（注明值冲突语义）
- [ ] 实现 `ToMap() map[K] V` 终止操作
- [ ] 与 US-005 的 `GroupBy`/`CountBy`/`PartitionBy` 返回值类型一致衔接
- [ ] 单测覆盖空 Map、键冲突、不可比键；通过 lint

### US-012: 集合运算（中间 / 终止操作）
**Description:** 作为使用者，我希望对两条流做并集、交集、差集。

**Acceptance Criteria:**
- [ ] 实现 `Union(other Seq[T])`、`Intersect(other Seq[T])`、`Difference(other Seq[T])`（`T comparable`）
- [ ] 实现顶层 `Union`/`Intersect`/`Difference` 多参数版本（参考 lo 的 `lo.Union` 等）
- [ ] 实现 `ContainsAll(other Seq[T]) bool` 终止操作
- [ ] godoc 注明实现选择（hash 去重 vs 有序），以及对内存的影响
- [ ] 单测覆盖交集为空、完全重叠、空流、不可比类型；通过 lint

### US-013: 并行操作（并发求值）
**Description:** 作为使用者，我希望对 CPU 密集型变换做并发求值以利用多核。

**Acceptance Criteria:**
- [ ] 实现 worker pool 基础设施（可配置 worker 数、有界缓冲、context 取消）
- [ ] 实现 `ParMap[U](func(T) U, opts ...Option) Seq[U]`、`ParFilter(func(T) bool, opts ...)`
- [ ] 提供保序（`WithOrder()`）与不保序（默认，更快）两种语义，godoc 注明
- [ ] 实现 `ParForEach`、`ParReduce`/`ParFold`（分片-合并）
- [ ] 并发安全：`-race` 下单测通过（`go test -race`）
- [ ] benchmark 证明多核下相对串行版本有加速（记录基准数据）
- [ ] 单测覆盖 panic 在 worker 内的传播、context 超时取消；通过 lint

### US-014: 函数式工具 helper（对标 lo 的通用工具）
**Description:** 作为使用者，我希望获得 lo 中常用的、与流无直接关系的通用工具函数，保持单一依赖来源。

**Acceptance Criteria:**
- [ ] 实现条件工具：`Ternary`、`Iff/IfElse` 链式（参考 lo `lo.If`）
- [ ] 实现默认值工具：`Coalesce`、`CoalesceOrZero`、`DefaultValue`、`Empty[T]()`
- [ ] 实现重试/容错：`Retry`、`RetryWithDelay`（与流可结合）
- [ ] 实现数值/约束 helper：`Clamp`、`Range`-related、`constraints` 约束别名（如 `Ordered`/`Number`）
- [ ] godoc 分类清晰；单测覆盖空指针、零值、边界；通过 lint

### US-015: 工程化交付（开源库标准）
**Description:** 作为开源库的维护者，我需要完整的工程化基础设施以保证质量与可维护性。

**Acceptance Criteria:**
- [ ] 所有公开符号附 godoc 注释，`go doc` / pkg.go.dev 渲染正常
- [ ] 每个公开方法提供 `Example*` 函数（`go test` 自动收集的 example），`go test ./...` 全绿
- [ ] 单测总覆盖率 ≥ 90%（`go test -cover`），关键路径 100%
- [ ] 提供关键操作的 benchmark（`Map`/`Filter`/`Reduce`/`ParMap` vs 串行 vs `slices` 直写 vs lo），`go test -bench` 可运行
- [ ] CI（GitHub Actions）：`go build`/`go vet`/`golangci-lint`/`go test -race`/coverage 上传，多版本矩阵（至少 `1.27`）
- [ ] 提供 `README.md`（含特性、安装、快速上手、与 lo 的对比表、性能说明）
- [ ] 提供 `LICENSE`、`CONTRIBUTING.md`、`CHANGELOG.md`，遵循 semver（v1 前用 v0.x）

## 4. Functional Requirements / 功能需求

- FR-1: 系统必须以 Go 1.27 为最低版本，并在 `go.mod` 中声明 `go 1.27`。
- FR-2: 系统必须提供具名包装类型 `Seq[T]` 与 `Seq2[K, V]`，其底层类型与 `iter.Seq` / `iter.Seq2` 二进制兼容、可互转。
- FR-3: 系统必须利用 Go 1.27 的范型方法特性，使方法可声明独立类型参数（如 `Map[U any]`）。
- FR-4: 所有**中间操作**（Map/Filter/Take/Drop/FlatMap/Chunk/SortBy/Zip 等）必须返回新的 `Seq`，且在无终止操作被调用前不执行任何计算（惰性）。
- FR-5: 中间操作必须支持**提前终止**：当下游 `yield` 返回 `false` 时，上游迭代必须立即停止。
- FR-6: 系统必须提供**终止操作**以触发求值：`Collect`/`ToSlice`/`ToMap`/`ToSet`/`Reduce`/`Fold`/`Count`/`ForEach`/`Any`/`All`/`Find`/`Min`/`Max` 等。
- FR-7: 系统必须为 `Seq2[K, V]` 提供与 `Seq[T]` 对等的核心链式操作集合。
- FR-8: 系统必须提供并行操作 `ParMap`/`ParFilter`/`ParForEach`/`ParReduce`，基于可配置的 worker pool，并支持保序与不保序两种模式。
- FR-9: 并行操作必须在 `go test -race` 下无数据竞争，且支持 context 取消与 worker 内 panic 的安全传播。
- FR-10: 系统必须对标 samber/lo，覆盖其集合类操作的主要类别（见 US 列表），并在 README 中给出与本库 API 的映射对照。
- FR-11: 系统对外不得引入任何第三方运行时依赖（仅依赖 Go 标准库；测试/benchmark/lint 工具不计入）。
- FR-12: 所有公开符号必须有 godoc 注释；每个公开方法必须有可编译运行的 `Example`。
- FR-13: 系统必须提供覆盖正常路径、边界、空输入的表驱动单测，公开包覆盖率 ≥ 90%。
- FR-14: 系统必须提供关键操作的 benchmark 与 CI 流水线（build/vet/lint/test -race/coverage）。

## 5. Non-Goals / 不在范围内

- **不**支持对**无界/无限流**的"自动惰性收集"——无限流只能配合 `Take`/`Find`/`Any` 等终止操作使用，`Collect` 无限流会阻塞/撑爆内存（godoc 注明）。
- **不**提供 mutable（可变）集合类型——本库是只读、不可变的流抽象。
- **不**追求与 samber/lo 的**逐函数 1:1 同名**；命名优先遵循 Scala 习惯（PascalCase 方法名），仅在工具函数上向 lo 看齐。
- **不**包含持久化、IO、网络、序列化能力。
- **不**提供字符串解析/正则、时间日期等与流无关的领域 helper（仅在 US-014 范围内的通用工具）。
- **不**支持 Go 1.27 以下的版本（范型方法是硬依赖）。
- **不**实现 R 式 dataframe / 列式存储等"大数据"语义。

## 6. Design Considerations / 设计考量

**核心类型骨架（示意）：**

```go
package seq

// Seq 是 iter.Seq[T] 的具名包装类型，可定义带类型参数的方法。
type Seq[T any] func(yield func(T) bool)

// Seq2 是 iter.Seq2[K,V] 的具名包装类型。
type Seq2[K any, V any] func(yield func(K, V) bool)

// 构造器
func Of[T any](vs ...T) Seq[T] { /* ... */ }

// 范型方法：方法自带类型参数 U，返回 Seq[U]，实现链式类型变换。
func (s Seq[T]) Map[U any](f func(T) U) Seq[U] {
    return func(yield func(U) bool) {
        for v := range s {
            if !yield(f(v)) { return }
        }
    }
}
```

**命名规范：**
- 类型：`Seq[T]`、`Seq2[K, V]`。
- 方法名：PascalCase，优先 Scala 词汇（`Map`/`Filter`/`FlatMap`/`Take`/`Drop`/`Collect`/`Fold`/`Reduce`/`GroupBy`/`Zip`/`ForEach`）。
- 顶层函数（非方法）：用于"多流输入"或无接收者的工具（`Zip`/`Unzip`/`Union`/`Range`/`Ternary`）。
- Option 风格的配置：`WithWorkers(n)`、`WithOrder()`、`WithContext(ctx)`。

**与 lo 的关系：** 在 README 中以对照表说明——lo 是 **eager + 函数式**（每次操作物化切片），本库是 **lazy + 方法链 + iter.Seq 原生**，二者可互补。

## 7. Technical Considerations / 技术考量

- **硬依赖 Go 1.27**：范型方法（type parameters on methods）是核心特性，`go.mod` 需 `go 1.27`；CI 需用 1.27 工具链（当前环境为 `go1.27rc1`）。
- **具名类型 vs 类型别名**：必须用具名类型 `type Seq[T any] func(yield func(T) bool)`（别名无法定义方法），需处理与 `iter.Seq` 之间的转换（底层类型相同，转换零成本）。
- **惰性 + break 语义**：所有迭代须遵循 `iter.Seq` 约定——`yield` 返回 `false` 即上游停止。终止操作是唯一触发求值的入口。
- **排序/去重/集合运算需缓冲**：`SortBy`/`Uniq`(全流)/`Union` 等无法纯惰性实现，需内部物化，必须在 godoc 明确标注其内存代价。
- **并行 worker pool**：需有界并发、context 取消、panic 捕获与重抛；保序模式需缓冲结果按索引回填。
- **测试可复现性**：`Shuffle`/`Sample` 注入 `*rand.Rand` 而非全局源；并行 benchmark 给出加速比基线。
- **API 兼容性**：v0.x 允许破坏性变更；v1 起承诺 semver，使用 `go mod` major suffix（`/v2`）策略。
- **零运行时依赖**：`go.mod` 的 `require` 块除自身外应为空（`go mod tidy` 后校验）。

## 8. Success Metrics / 成功指标

- 公开 API 覆盖 lo 主要集合类别 ≥ 90%（以 README 对照表中"已实现"项计）。
- 单测覆盖率公开包 ≥ 90%，关键路径（惰性求值、提前终止、并发安全）100%。
- `go test -race ./...`、`go vet ./...`、`golangci-lint run` 全绿。
- 在代表性 benchmark 上，惰性链式相对"多次切片直写"减少 ≥ 50% 的分配次数（以 `benchmem` 记录）。
- `ParMap` 在 ≥ 4 核机器上相对串行 `Map`（CPU 密集变换）取得可测加速（记录加速比）。
- pkg.go.dev 文档与所有 `Example` 正常渲染、可运行。

## 9. Open Questions / 待确认问题

1. **Module path**：建议 `github.com/lzr/seq`（基于当前目录 `/root/work/lzr/seq`），需确认 GitHub 仓库归属与最终 import path。
2. **包名**：默认 `seq`。是否需要顶层拆分为子包（如 `seq`/`seq2`/`seq/parallel`）还是单包扁平结构？（建议 v1 单包，降低链式调用时的 import 成本）
3. **`Compact` 语义**：参照 lo 去除零值，还是去除"相邻重复"？需在实现前明确（影响命名与 godoc）。
4. **`ToMap` 键冲突策略**：覆盖 / 报错 / 保留首个 / 保留最后？建议默认覆盖 + 提供 `ToMapWith(merge)` 变体。
5. **`Range` 是否泛型数值**：仅 `int`，还是泛型支持 `int/float` 区间（需约束包）？建议 v1 仅 `int`。
6. **`ParMap` 默认并发度**：`runtime.NumCPU()` 还是 `GOMAXPROCS(0)`？保序是否作为默认？
7. **`SortBy` 缓冲语义**：返回 `Seq[T]`（延迟到终止时排序）还是返回 `[]T`？建议返回 `Seq[T]` 以保持链式一致性，但 godoc 标注缓冲代价。
8. **命名风格最终裁决**：Scala 风格（`FlatMap`/`GroupBy`/`ForEach`）vs lo 风格（`Flatmap` 已无，但 `lo.GroupBy` 等）vs Go 惯例——本 PRD 默认 Scala 风格，需最终确认。
