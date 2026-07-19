// 同城车辆买卖模块 API 封装（完整版 v1.0）
// 对应后端路由前缀：/api/v1/car（公开/C端） + /api/v1/admin/car（管理后台，注意 admin 在 car 之前）
// 涵盖：车源/发布单/检测/评估/分期/车险/试驾/过户/合同/担保/评价/举报/统计/推荐/审核规则/车型库
import request from '@/utils/request'

// ====================================================================
// 一、车源管理（公开 + C 端登录 + 管理后台）
// ====================================================================

// --- 公开接口（无需登录） ---

// 车源列表
export function listCars(params) {
  return request.get('/car', { params })
}

// 搜索车源
export function searchCars(params) {
  return request.get('/car/search', { params })
}

// 附近车源
export function listNearbyCars(params) {
  return request.get('/car/nearby', { params })
}

// 高级搜索
export function advancedSearchCars(params) {
  return request.get('/car/advanced-search', { params })
}

// 车源详情
export function getCar(id) {
  return request.get(`/car/${id}`)
}

// 收藏状态（公开）
export function getCarFavStatus(id) {
  return request.get(`/car/${id}/fav`)
}

// --- C 端登录接口 ---

// 我的车源
export function listMyCars(params) {
  return request.get('/car/mine', { params })
}

// 我的收藏
export function listMyCarFavs(params) {
  return request.get('/car/favorites', { params })
}

// 发布车源
export function createCar(data) {
  return request.post('/car', data)
}

// 更新车源
export function updateCar(id, data) {
  return request.put(`/car/${id}`, data)
}

// 删除车源
export function deleteCar(id) {
  return request.delete(`/car/${id}`)
}

// 收藏车源
export function favCar(id) {
  return request.post(`/car/${id}/fav`)
}

// 联系次数
export function incrCarContact(id) {
  return request.post(`/car/${id}/contact`)
}

// 分享次数
export function incrCarShare(id) {
  return request.post(`/car/${id}/share`)
}

// 记录浏览
export function recordCarView(id) {
  return request.post(`/car/${id}/views`)
}

// --- 管理后台接口（需 car:audit 权限） ---

// 管理端 - 车源列表
export function adminListCars(params) {
  return request.get('/admin/car/cars', { params })
}

// 管理端 - 车源详情
export function adminGetCar(id) {
  return request.get(`/admin/car/cars/${id}`)
}

// 管理端 - 审核
export function auditCar(id, data) {
  return request.put(`/admin/car/cars/${id}/audit`, data)
}

// 管理端 - 状态变更
export function adminUpdateCarStatus(id, data) {
  return request.put(`/admin/car/cars/${id}/status`, data)
}

// 管理端 - 真车验证
export function realCarVerify(id, data) {
  return request.put(`/admin/car/cars/${id}/real-car-verify`, data)
}

// 管理端 - 推广
export function updateCarPromotion(id, data) {
  return request.put(`/admin/car/cars/${id}/promotion`, data)
}

// ====================================================================
// 二、发布单管理（listing）
// ====================================================================

// 发布单列表（公开）
export function listListings(params) {
  return request.get('/car/listings', { params })
}

// 发布单详情（公开）
export function getListing(id) {
  return request.get(`/car/listings/${id}`)
}

// 创建发布单
export function createListing(data) {
  return request.post('/car/listings', data)
}

// 更新发布单
export function updateListing(id, data) {
  return request.put(`/car/listings/${id}`, data)
}

// 删除发布单
export function deleteListing(id) {
  return request.delete(`/car/listings/${id}`)
}

// 我的发布单
export function listMyListings(params) {
  return request.get('/car/listings/mine', { params })
}

// 管理端 - 发布单列表
export function adminListListings(params) {
  return request.get('/admin/car/listings', { params })
}

// 管理端 - 发布单详情
export function adminGetListing(id) {
  return request.get(`/admin/car/listings/${id}`)
}

// 管理端 - 发布单审核
export function auditListing(id, data) {
  return request.put(`/admin/car/listings/${id}/audit`, data)
}

// 管理端 - 发布单状态变更
export function adminUpdateListingStatus(id, data) {
  return request.put(`/admin/car/listings/${id}/status`, data)
}

// 管理端 - 检测状态变更
export function updateListingInspectionStatus(id, data) {
  return request.put(`/admin/car/listings/${id}/inspection-status`, data)
}

// ====================================================================
// 三、检测报告（inspection）
// ====================================================================

// 检测列表（公开）
export function listInspections(params) {
  return request.get('/car/inspections', { params })
}

// 检测详情（公开）
export function getInspection(id) {
  return request.get(`/car/inspections/${id}`)
}

// 按车源查检测
export function getInspectionByCar(carId) {
  return request.get(`/car/cars/${carId}/inspection`)
}

// 创建检测
export function createInspection(data) {
  return request.post('/car/inspections', data)
}

// 更新检测
export function updateInspection(id, data) {
  return request.put(`/car/inspections/${id}`, data)
}

// 删除检测
export function deleteInspection(id) {
  return request.delete(`/car/inspections/${id}`)
}

// 我的检测
export function listMyInspections(params) {
  return request.get('/car/inspections/mine', { params })
}

// 管理端 - 检测列表
export function adminListInspections(params) {
  return request.get('/admin/car/inspections', { params })
}

// 管理端 - 检测详情
export function adminGetInspection(id) {
  return request.get(`/admin/car/inspections/${id}`)
}

// 管理端 - 检测审核
export function reviewInspection(id, data) {
  return request.put(`/admin/car/inspections/${id}/review`, data)
}

// 管理端 - 检测状态变更
export function adminUpdateInspectionStatus(id, data) {
  return request.put(`/admin/car/inspections/${id}/status`, data)
}

// ====================================================================
// 四、车辆评估（evaluation）
// ====================================================================

// 评估列表（公开）
export function listEvaluations(params) {
  return request.get('/car/evaluations', { params })
}

// 评估详情（公开）
export function getEvaluation(id) {
  return request.get(`/car/evaluations/${id}`)
}

// 按车源查评估
export function getEvaluationByCar(carId) {
  return request.get(`/car/cars/${carId}/evaluation`)
}

// 按车源查评估列表
export function listEvaluationsByCar(carId, params) {
  return request.get(`/car/cars/${carId}/evaluations`, { params })
}

// 创建评估
export function createEvaluation(data) {
  return request.post('/car/evaluations', data)
}

// 更新评估
export function updateEvaluation(id, data) {
  return request.put(`/car/evaluations/${id}`, data)
}

// 删除评估
export function deleteEvaluation(id) {
  return request.delete(`/car/evaluations/${id}`)
}

// 我的评估
export function listMyEvaluations(params) {
  return request.get('/car/evaluations/mine', { params })
}

// 在线估值
export function onlineEvaluate(data) {
  return request.post('/car/evaluations/online', data)
}

// 管理端 - 评估列表
export function adminListEvaluations(params) {
  return request.get('/admin/car/evaluations', { params })
}

// 管理端 - 评估详情
export function adminGetEvaluation(id) {
  return request.get(`/admin/car/evaluations/${id}`)
}

// 管理端 - 评估状态变更
export function adminUpdateEvaluationStatus(id, data) {
  return request.put(`/admin/car/evaluations/${id}/status`, data)
}

// ====================================================================
// 五、分期付款（financing）
// ====================================================================

// 分期列表（公开，仅已发布）
export function listFinancings(params) {
  return request.get('/car/financings', { params })
}

// 热门分期
export function listHotFinancings(params) {
  return request.get('/car/financings/hot', { params })
}

// 分期详情
export function getFinancing(id) {
  return request.get(`/car/financings/${id}`)
}

// 分期计算
export function calculateFinancing(data) {
  return request.post('/car/financings/calculate', data)
}

// 管理端 - 分期列表
export function adminListFinancings(params) {
  return request.get('/admin/car/financings', { params })
}

// 管理端 - 分期详情
export function adminGetFinancing(id) {
  return request.get(`/admin/car/financings/${id}`)
}

// 管理端 - 创建分期
export function createFinancing(data) {
  return request.post('/admin/car/financings', data)
}

// 管理端 - 更新分期
export function updateFinancing(id, data) {
  return request.put(`/admin/car/financings/${id}`, data)
}

// 管理端 - 删除分期
export function deleteFinancing(id) {
  return request.delete(`/admin/car/financings/${id}`)
}

// 管理端 - 分期状态变更
export function adminUpdateFinancingStatus(id, data) {
  return request.put(`/admin/car/financings/${id}/status`, data)
}

// ====================================================================
// 六、车险管理（insurance）
// ====================================================================

// 车险列表（公开，仅已发布）
export function listInsurances(params) {
  return request.get('/car/insurances', { params })
}

// 热门车险
export function listHotInsurances(params) {
  return request.get('/car/insurances/hot', { params })
}

// 车险详情
export function getInsurance(id) {
  return request.get(`/car/insurances/${id}`)
}

// 车险报价
export function quoteInsurance(data) {
  return request.post('/car/insurances/quote', data)
}

// 管理端 - 车险列表
export function adminListInsurances(params) {
  return request.get('/admin/car/insurances', { params })
}

// 管理端 - 车险详情
export function adminGetInsurance(id) {
  return request.get(`/admin/car/insurances/${id}`)
}

// 管理端 - 创建车险
export function createInsurance(data) {
  return request.post('/admin/car/insurances', data)
}

// 管理端 - 更新车险
export function updateInsurance(id, data) {
  return request.put(`/admin/car/insurances/${id}`, data)
}

// 管理端 - 删除车险
export function deleteInsurance(id) {
  return request.delete(`/admin/car/insurances/${id}`)
}

// 管理端 - 车险状态变更
export function adminUpdateInsuranceStatus(id, data) {
  return request.put(`/admin/car/insurances/${id}/status`, data)
}

// ====================================================================
// 七、试驾预约（test-drive）
// ====================================================================

// 试驾列表（公开）
export function listTestDrives(params) {
  return request.get('/car/test-drives', { params })
}

// 试驾详情（公开）
export function getTestDrive(id) {
  return request.get(`/car/test-drives/${id}`)
}

// 按经销商查试驾
export function listTestDrivesByDealer(dealerId, params) {
  return request.get(`/car/dealers/${dealerId}/test-drives`, { params })
}

// 创建试驾预约
export function createTestDrive(data) {
  return request.post('/car/test-drives', data)
}

// 更新试驾
export function updateTestDrive(id, data) {
  return request.put(`/car/test-drives/${id}`, data)
}

// 取消试驾
export function cancelTestDrive(id) {
  return request.post(`/car/test-drives/${id}/cancel`)
}

// 我的试驾
export function listMyTestDrives(params) {
  return request.get('/car/test-drives/mine', { params })
}

// 销售试驾
export function listSalesTestDrives(params) {
  return request.get('/car/test-drives/sales', { params })
}

// 上传驾照
export function uploadTestDriveLicense(id, data) {
  return request.post(`/car/test-drives/${id}/license`, data)
}

// 管理端 - 试驾列表
export function adminListTestDrives(params) {
  return request.get('/admin/car/test-drives', { params })
}

// 管理端 - 试驾详情
export function adminGetTestDrive(id) {
  return request.get(`/admin/car/test-drives/${id}`)
}

// 管理端 - 试驾状态变更
export function adminUpdateTestDriveStatus(id, data) {
  return request.put(`/admin/car/test-drives/${id}/status`, data)
}

// ====================================================================
// 八、过户办理（transfer）
// ====================================================================

// 过户详情（公开）
export function getTransfer(id) {
  return request.get(`/car/transfers/${id}`)
}

// 按车源查过户
export function getTransferByCar(carId) {
  return request.get(`/car/cars/${carId}/transfer`)
}

// 创建过户
export function createTransfer(data) {
  return request.post('/car/transfers', data)
}

// 更新过户
export function updateTransfer(id, data) {
  return request.put(`/car/transfers/${id}`, data)
}

// 删除过户
export function deleteTransfer(id) {
  return request.delete(`/car/transfers/${id}`)
}

// 过户列表
export function listTransfers(params) {
  return request.get('/car/transfers', { params })
}

// 我卖出的过户
export function listSoldTransfers(params) {
  return request.get('/car/transfers/sold', { params })
}

// 我买进的过户
export function listBoughtTransfers(params) {
  return request.get('/car/transfers/bought', { params })
}

// 管理端 - 过户列表
export function adminListTransfers(params) {
  return request.get('/admin/car/transfers', { params })
}

// 管理端 - 过户详情
export function adminGetTransfer(id) {
  return request.get(`/admin/car/transfers/${id}`)
}

// 管理端 - 过户状态变更
export function adminUpdateTransferStatus(id, data) {
  return request.put(`/admin/car/transfers/${id}/status`, data)
}

// ====================================================================
// 九、担保交易（escrow）
// ====================================================================

// 担保列表
export function listEscrows(params) {
  return request.get('/car/escrows', { params })
}

// 担保详情
export function getEscrow(id) {
  return request.get(`/car/escrows/${id}`)
}

// 按担保编号查
export function getEscrowByNo(escrowNo) {
  return request.get(`/car/escrows/no/${escrowNo}`)
}

// 按车源查担保
export function listEscrowsByCar(carId, params) {
  return request.get(`/car/cars/${carId}/escrows`, { params })
}

// 担保动作（pay/release/refund/dispute/cancel）
export function escrowAction(id, data) {
  return request.post(`/car/escrows/${id}/action`, data)
}

// 管理端 - 担保列表
export function adminListEscrows(params) {
  return request.get('/admin/car/escrows', { params })
}

// 管理端 - 担保详情
export function adminGetEscrow(id) {
  return request.get(`/admin/car/escrows/${id}`)
}

// 管理端 - 担保状态变更
export function adminUpdateEscrowStatus(id, data) {
  return request.put(`/admin/car/escrows/${id}/status`, data)
}

// ====================================================================
// 十、合同管理（contract）
// ====================================================================

// 合同列表
export function listContracts(params) {
  return request.get('/car/contracts', { params })
}

// 合同详情
export function getContract(id) {
  return request.get(`/car/contracts/${id}`)
}

// 按合同编号查
export function getContractByNo(contractNo) {
  return request.get(`/car/contracts/no/${contractNo}`)
}

// 按车源查合同
export function listContractsByCar(carId, params) {
  return request.get(`/car/cars/${carId}/contracts`, { params })
}

// 签署合同
export function signContract(id) {
  return request.post(`/car/contracts/${id}/sign`)
}

// 终止合同
export function terminateContract(id) {
  return request.post(`/car/contracts/${id}/terminate`)
}

// 管理端 - 合同列表
export function adminListContracts(params) {
  return request.get('/admin/car/contracts', { params })
}

// 管理端 - 合同详情
export function adminGetContract(id) {
  return request.get(`/admin/car/contracts/${id}`)
}

// 管理端 - 合同状态变更
export function adminUpdateContractStatus(id, data) {
  return request.put(`/admin/car/contracts/${id}/status`, data)
}

// ====================================================================
// 十一、评价管理（review）
// ====================================================================

// 评价列表（公开）
export function listReviews(params) {
  return request.get('/car/reviews', { params })
}

// 按目标查评价
export function listReviewsByTarget(params) {
  return request.get('/car/reviews/by-target', { params })
}

// 评价统计
export function getReviewStats(params) {
  return request.get('/car/reviews/stats', { params })
}

// 评价详情
export function getReview(id) {
  return request.get(`/car/reviews/${id}`)
}

// 我的评价
export function listMyReviews(params) {
  return request.get('/car/reviews/mine', { params })
}

// 创建评价
export function createReview(data) {
  return request.post('/car/reviews', data)
}

// 更新评价
export function updateReview(id, data) {
  return request.put(`/car/reviews/${id}`, data)
}

// 删除评价
export function deleteReview(id) {
  return request.delete(`/car/reviews/${id}`)
}

// 评价回复
export function replyReview(id, data) {
  return request.post(`/car/reviews/${id}/reply`, data)
}

// 评价追加
export function appendReview(id, data) {
  return request.post(`/car/reviews/${id}/append`, data)
}

// 评价点赞
export function likeReview(id) {
  return request.post(`/car/reviews/${id}/like`)
}

// 管理端 - 评价列表
export function adminListReviews(params) {
  return request.get('/admin/car/reviews', { params })
}

// 管理端 - 评价状态变更
export function adminUpdateReviewStatus(id, data) {
  return request.put(`/admin/car/reviews/${id}/status`, data)
}

// ====================================================================
// 十二、举报管理（report）
// ====================================================================

// 按目标查举报
export function listReportsByTarget(params) {
  return request.get('/car/reports/by-target', { params })
}

// 创建举报
export function createReport(data) {
  return request.post('/car/reports', data)
}

// 我的举报
export function listMyReports(params) {
  return request.get('/car/reports/mine', { params })
}

// 举报申诉
export function appealReport(id, data) {
  return request.post(`/car/reports/${id}/appeal`, data)
}

// 管理端 - 举报列表
export function adminListReports(params) {
  return request.get('/admin/car/reports', { params })
}

// 管理端 - 举报详情
export function adminGetReport(id) {
  return request.get(`/admin/car/reports/${id}`)
}

// 管理端 - 处理举报
export function processReport(id, data) {
  return request.put(`/admin/car/reports/${id}/process`, data)
}

// 管理端 - 申诉处理
export function processAppeal(id, data) {
  return request.put(`/admin/car/reports/${id}/appeal`, data)
}

// 管理端 - 举报状态变更
export function adminUpdateReportStatus(id, data) {
  return request.put(`/admin/car/reports/${id}/status`, data)
}

// ====================================================================
// 十三、推荐管理（recommendation）
// ====================================================================

// 按用户查推荐
export function listRecommendations(params) {
  return request.get('/car/recommendations', { params })
}

// 按车源查推荐
export function listRecommendationsByCar(carId, params) {
  return request.get(`/car/cars/${carId}/recommendations`, { params })
}

// 标记已点击
export function markRecClicked(id) {
  return request.post(`/car/recommendations/${id}/click`)
}

// 标记已联系
export function markRecContacted(id) {
  return request.post(`/car/recommendations/${id}/contact`)
}

// 标记已忽略
export function markRecDismissed(id) {
  return request.post(`/car/recommendations/${id}/dismiss`)
}

// 管理端 - 推荐列表
export function adminListRecommendations(params) {
  return request.get('/admin/car/recommendations', { params })
}

// 管理端 - 删除推荐
export function adminDeleteRecommendation(id) {
  return request.delete(`/admin/car/recommendations/${id}`)
}

// ====================================================================
// 十四、审核规则（audit-rule）
// ====================================================================

// 管理端 - 规则列表
export function listAuditRules(params) {
  return request.get('/admin/car/audit-rules', { params })
}

// 管理端 - 规则详情
export function getAuditRule(id) {
  return request.get(`/admin/car/audit-rules/${id}`)
}

// 管理端 - 创建规则
export function createAuditRule(data) {
  return request.post('/admin/car/audit-rules', data)
}

// 管理端 - 更新规则
export function updateAuditRule(id, data) {
  return request.put(`/admin/car/audit-rules/${id}`, data)
}

// 管理端 - 删除规则
export function deleteAuditRule(id) {
  return request.delete(`/admin/car/audit-rules/${id}`)
}

// 管理端 - 规则状态变更
export function updateAuditRuleStatus(id, data) {
  return request.put(`/admin/car/audit-rules/${id}/status`, data)
}

// 管理端 - 规则检查
export function checkAuditRules(data) {
  return request.post('/admin/car/audit-rules/check', data)
}

// ====================================================================
// 十五、车型库（catalog: 品牌/系列/分类）
// ====================================================================

// 车型列表（公开）
export function listModels(params) {
  return request.get('/car/models', { params })
}

// 车型详情
export function getModel(id) {
  return request.get(`/car/models/${id}`)
}

// 品牌列表（公开）
export function listBrands(params) {
  return request.get('/car/brands', { params })
}

// 全部品牌
export function listAllBrands() {
  return request.get('/car/brands/all')
}

// 按品牌查车型
export function listModelsByBrand(brandId, params) {
  return request.get(`/car/brands/${brandId}/models`, { params })
}

// 分类列表（公开）
export function listCategories(params) {
  return request.get('/car/categories', { params })
}

// 按层级查分类
export function listCategoriesByLevel(level, params) {
  return request.get(`/car/categories/level/${level}`, { params })
}

// 分类详情
export function getCategory(id) {
  return request.get(`/car/categories/${id}`)
}

// 子分类
export function listCategoriesByParent(id, params) {
  return request.get(`/car/categories/${id}/children`, { params })
}

// ====================================================================
// 十六、数据统计
// ====================================================================

// 公开 - 热门车源
export function getHotCars(params) {
  return request.get('/car/statistics/hot-cars', { params })
}

// 公开 - 价格趋势
export function getPriceTrend(params) {
  return request.get('/car/statistics/price-trend', { params })
}

// C 端 - 卖家统计
export function getSellerStats() {
  return request.get('/car/statistics/seller')
}

// 管理端 - 平台总览
export function getOverviewStats() {
  return request.get('/admin/car/statistics/overview')
}

// 管理端 - 统计列表
export function adminListStats(params) {
  return request.get('/admin/car/statistics', { params })
}

// ====================================================================
// 兼容别名
// ====================================================================

export const adminListCarReports = adminListReports
export const adminListCarReviews = adminListReviews
export const getCarDetail = adminGetCar
export const getOverviewStatsCar = getOverviewStats
