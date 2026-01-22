# 主节点前端页面规划（Nginx Pod 展示/管理集群）

> 目的：在主节点部署一个 Nginx Pod 承载静态前端页面，用于集群可视化与任务管理。

## 1. 页面目标与范围
- 展示集群概览与任务编排状态（Mission/MissionStage/FlightTask/Weapon）。
- 提供常用只读视图 + 少量受控操作入口（如：触发示例任务、查看 Pod 详情、下载日志）。
- 低耦合：前端为静态站点，业务数据通过 API 获取，不在 Nginx 内做逻辑。

## 2. 信息架构（IA）
- 首页 / Dashboard
  - 集群概况：节点数量、Ready/NotReady、版本、核心插件状态
  - 任务总览：Mission/Stage/Task 数量与状态分布
  - 近 24h 事件摘要：FailedScheduling、ImagePullBackOff 等
- 任务视图
  - Mission 列表 → Mission 详情 → Stage 详情 → FlightTask 详情
  - 状态流转时间线（Pending → Scheduled → Running → Succeeded/Failed）
  - 依赖关系视图（Stage 依赖与并行/串行策略）
- FlightTask 视图
  - Pod 绑定情况、调度约束解析、Weapon sidecar 注入情况
  - 条件与调度信息（PodScheduled、FailedScheduling 计数）
- Weapon 视图
  - 兼容性配置、镜像版本、使用情况
- 集群资源视图（轻量）
  - Node/Pod/Namespace/Events（只读）
- 设置与帮助
  - API 连接状态、刷新频率、权限说明

## 3. 交互与功能规划
- 只读优先：默认不提供直接写操作，避免误操作。
- 可选受控操作（后续）
  - 触发示例任务/删除 CRD（需要鉴权）
  - 复制 YAML/导出 JSON
  - 跳转到 kubectl 命令提示

## 4. 技术方案（前端 + 数据）
- 前端框架：Vite + Vue3 构建 SPA，产物静态文件交给 Nginx。
- 数据来源：
  - 优先：自建轻量 API（聚合 CRD + 原生资源），供前端调用。
  - 备选：直接调用 Kubernetes API（需做鉴权与 CORS 处理）。
- 数据刷新策略：轮询 + 手动刷新，默认 10s；任务详情可局部刷新 3–5s。
- 版本与环境配置：使用 `window.__APP_CONFIG__` 注入 API_BASE。

## 5. 视觉与布局方向
- 风格：偏“作战态势板”，深色中性基调 + 明确的状态色。
- 排版：
  - 标题/数字采用强识别字体（如 "IBM Plex Sans Condensed" 或 "Rajdhani"）。
  - 正文采用可读性强的无衬线体。
- 颜色：
  - 主色：深灰蓝；强调色：橙/青；状态色：绿/黄/红/灰。
  - 避免紫色默认风格与大面积纯白背景。
- 动效：
  - 页面加载淡入 + 卡片分层上浮
  - 列表逐项出现（staggered）
- 背景：低对比渐变 + 微弱网格/噪点纹理。

## 6. 页面组件拆分
- 顶部：全局状态条（API 状态、刷新按钮、时间）。
- 左侧：导航。
- 中央：核心数据视图（卡片/表格/时间线/关系图）。
- 右侧：详情面板（选中任务/Pod/武器的详细信息）。

## 7. 安全与权限
- 只读模式默认启用；写操作需明确鉴权与确认。
- API 层可根据 ServiceAccount + RBAC 限制。
- 前端避免直接持久化敏感信息。

## 8. 部署结构建议
- Nginx Pod（主节点）：
  - `ConfigMap` 挂载静态资源
  - `Service` + `Ingress`（或 NodePort）暴露
- API 服务（建议）
  - 独立 Deployment（聚合 CRD + 只读 K8s 资源）
  - 为前端提供统一 API

## 9. 最小可用版本（MVP）
- Dashboard 基础卡片（集群/任务统计）
- Mission/Stage/FlightTask 列表与详情
- FlightTask → Pod 状态展示
- Weapon 列表与详情

## 10. 后续增强方向
- 依赖/阶段甘特图
- 自定义调度评分可视化
- 任务回放与审计日志
- 多集群切换

## 11. 前端任务列表（Vue3）

### 11.1 需求与接口确认
- [ ] 明确前端访问路径（聚合 API 或直接 K8s API）
- [ ] 列出首批页面所需字段与查询条件
- [ ] 定义 API_BASE 与环境配置策略（dev/stage/prod）

### 11.2 工程初始化
- [ ] 创建 Vite + Vue3 项目
- [ ] 引入 Vue Router + Pinia + Axios（或 Fetch 封装）
- [ ] 约定目录结构（pages/components/store/api/styles）
- [ ] 配置基础构建与本地开发脚本

### 11.3 视觉系统与布局骨架
- [ ] 定义 CSS 变量（颜色、间距、阴影、边框、状态色）
- [ ] 选择标题/正文两套字体并设定全局排版
- [ ] 搭建全局布局：Header / Sidebar / Main / Detail Panel

### 11.4 组件库（最小集）
- [ ] 状态标签（Pending/Scheduled/Running/Succeeded/Failed）
- [ ] KPI 卡片（数字 + 趋势/占比）
- [ ] 事件列表（按时间排序）
- [ ] 关系与时间线组件（任务推进）
- [ ] Key-Value 详情面板
- [ ] 数据表格（可排序/筛选）

### 11.5 页面开发（MVP）
- [ ] Dashboard：集群概况 + 任务统计 + 事件摘要
- [ ] Mission 列表与详情（阶段进度与依赖）
- [ ] MissionStage 详情（串并行策略 + FlightTask 列表）
- [ ] FlightTask 详情（Pod/调度/sidecar/条件）
- [ ] Weapon 列表与详情（兼容性与版本）
- [ ] 资源视图（Nodes/Pods/Events 只读）
- [ ] 设置页（API 状态、刷新频率、帮助）

### 11.6 数据层与状态管理
- [ ] API 客户端封装（统一错误处理与重试）
- [ ] 定义 CRD 类型与响应结构
- [ ] 实现轮询 + 手动刷新机制
- [ ] 空态/异常态/加载态 UI

### 11.7 构建与部署
- [ ] 生产构建输出到 `dist/`
- [ ] Nginx SPA 路由回退配置（`try_files`）
- [ ] K8s 部署清单：ConfigMap + Deployment + Service + Ingress
- [ ] 节点调度：主节点 nodeSelector + tolerations（如需）

### 11.8 验证与验收
- [ ] 使用示例 CRD 进行页面联调
- [ ] 移动端/桌面端布局检查
- [ ] 基础性能与可读性自检
