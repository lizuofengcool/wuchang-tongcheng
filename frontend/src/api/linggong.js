// 同城零工兼职模块 API 封装
// 对应后端路由前缀：/api/v1/linggong
// 涵盖：岗位/任务/报名/雇主/工人/合同/支付/评价/技能/资质/信用/审核规则
import request from '@/utils/request'

// ====================================================================
// 一、岗位主表（linggongs）
// ====================================================================

// 管理端列表（分页 + 筛选）
export function adminListLinggongs(params) {
  return request.get('/linggong/admin/linggongs', { params })
}

// 管理端详情
export function adminGetLinggong(id) {
  return request.get(`/linggong/admin/linggongs/${id}`)
}

// 公开列表
export function listLinggongs(params) {
  return request.get('/linggong', { params })
}

// 公开详情
export function getLinggong(id) {
  return request.get(`/linggong/${id}`)
}

// 搜索
export function searchLinggongs(params) {
  return request.get('/linggong/search', { params })
}

// 附近
export function listNearbyLinggongs(params) {
  return request.get('/linggong/nearby', { params })
}

// 创建岗位（C 端登录）
export function createLinggong(data) {
  return request.post('/linggong', data)
}

// 更新岗位
export function updateLinggong(id, data) {
  return request.put(`/linggong/${id}`, data)
}

// 删除岗位
export function deleteLinggong(id) {
  return request.delete(`/linggong/${id}`)
}

// 我的发布
export function listMyLinggongs(params) {
  return request.get('/linggong/mine', { params })
}

// 雇主下岗位
export function listLinggongsByEmployer(employerId, params) {
  return request.get(`/linggong/employers/${employerId}/linggongs`, { params })
}

// 审核（audit_status: 0待审 1通过 2拒绝）
export function auditLinggong(id, data) {
  return request.put(`/linggong/admin/linggongs/${id}/audit`, data)
}

// 状态变更（status: 1发布 2下架 3过期 5满员 6关闭 7完成）
export function adminUpdateLinggongStatus(id, status) {
  return request.put(`/linggong/admin/linggongs/${id}/status`, { status })
}

// 批量状态变更
export function batchUpdateLinggongStatus(data) {
  return request.post('/linggong/admin/linggongs/batch-status', data)
}

// 浏览/联系/分享 自增
export function incrLinggongView(id) {
  return request.post(`/linggong/${id}/view`)
}
export function incrLinggongContact(id) {
  return request.post(`/linggong/${id}/contact`)
}
export function incrLinggongShare(id) {
  return request.post(`/linggong/${id}/share`)
}

// ====================================================================
// 二、任务包（linggong_tasks）
// ====================================================================

export function listLinggongTasks(params) {
  return request.get('/linggong/tasks', { params })
}
export function getLinggongTask(id) {
  return request.get(`/linggong/tasks/${id}`)
}
export function createLinggongTask(data) {
  return request.post('/linggong/tasks', data)
}
export function updateLinggongTask(id, data) {
  return request.put(`/linggong/tasks/${id}`, data)
}
export function deleteLinggongTask(id) {
  return request.delete(`/linggong/tasks/${id}`)
}
export function listTasksByEmployer(employerId, params) {
  return request.get(`/linggong/tasks/employer/${employerId}`, { params })
}
export function listTasksByLinggong(linggongId) {
  return request.get(`/linggong/${linggongId}/tasks`)
}
export function claimLinggongTask(id, data) {
  return request.post(`/linggong/tasks/${id}/claim`, data)
}
export function submitLinggongTask(id, data) {
  return request.post(`/linggong/tasks/${id}/submit`, data)
}
export function verifyLinggongTask(id, data) {
  return request.post(`/linggong/tasks/${id}/verify`, data)
}

// 管理端
export function adminListLinggongTasks(params) {
  return request.get('/linggong/admin/tasks', { params })
}
export function adminGetLinggongTask(id) {
  return request.get(`/linggong/admin/tasks/${id}`)
}
export function adminUpdateLinggongTaskStatus(id, status) {
  return request.put(`/linggong/admin/tasks/${id}/status`, { status })
}

// ====================================================================
// 三、报名申请（linggong_applications）
// ====================================================================

export function listLinggongApplications(params) {
  return request.get('/linggong/applications', { params })
}
export function getLinggongApplication(id) {
  return request.get(`/linggong/applications/${id}`)
}
export function createLinggongApplication(data) {
  return request.post('/linggong/applications', data)
}
export function updateLinggongApplication(id, data) {
  return request.put(`/linggong/applications/${id}`, data)
}
export function deleteLinggongApplication(id) {
  return request.delete(`/linggong/applications/${id}`)
}
export function auditLinggongApplication(id, data) {
  return request.put(`/linggong/applications/${id}/audit`, data)
}
export function cancelLinggongApplication(id) {
  return request.post(`/linggong/applications/${id}/cancel`)
}
export function listApplicationsByLinggong(linggongId, params) {
  return request.get(`/linggong/${linggongId}/applications`, { params })
}
export function listMyApplications(params) {
  return request.get('/linggong/applications/mine', { params })
}
export function listEmployerApplications(params) {
  return request.get('/linggong/applications/employer', { params })
}

// 管理端
export function adminListLinggongApplications(params) {
  return request.get('/linggong/admin/applications', { params })
}

// ====================================================================
// 四、雇主认证（linggong_employers）
// ====================================================================

export function listLinggongEmployers(params) {
  return request.get('/linggong/employers', { params })
}
export function getLinggongEmployer(id) {
  return request.get(`/linggong/employers/${id}`)
}
export function createLinggongEmployer(data) {
  return request.post('/linggong/employers', data)
}
export function updateLinggongEmployer(id, data) {
  return request.put(`/linggong/employers/${id}`, data)
}
export function deleteLinggongEmployer(id) {
  return request.delete(`/linggong/employers/${id}`)
}
export function getMyLinggongEmployer() {
  return request.get('/linggong/employers/me')
}

// 管理端
export function adminListLinggongEmployers(params) {
  return request.get('/linggong/admin/employers', { params })
}
export function auditLinggongEmployer(id, data) {
  return request.put(`/linggong/admin/employers/${id}/audit`, data)
}
export function adminUpdateLinggongEmployerStatus(id, status) {
  return request.put(`/linggong/admin/employers/${id}/status`, { status })
}

// ====================================================================
// 五、求职者档案（linggong_workers）
// ====================================================================

export function listLinggongWorkers(params) {
  return request.get('/linggong/workers', { params })
}
export function getLinggongWorker(id) {
  return request.get(`/linggong/workers/${id}`)
}
export function createLinggongWorker(data) {
  return request.post('/linggong/workers', data)
}
export function updateLinggongWorker(id, data) {
  return request.put(`/linggong/workers/${id}`, data)
}
export function deleteLinggongWorker(id) {
  return request.delete(`/linggong/workers/${id}`)
}
export function getMyLinggongWorker() {
  return request.get('/linggong/workers/me')
}

// 管理端
export function adminListLinggongWorkers(params) {
  return request.get('/linggong/admin/workers', { params })
}
export function auditLinggongWorker(id, data) {
  return request.put(`/linggong/admin/workers/${id}/audit`, data)
}
export function adminUpdateLinggongWorkerStatus(id, status) {
  return request.put(`/linggong/admin/workers/${id}/status`, { status })
}

// ====================================================================
// 六、合同（linggong_contracts）
// ====================================================================

export function listLinggongContracts(params) {
  return request.get('/linggong/contracts', { params })
}
export function getLinggongContract(id) {
  return request.get(`/linggong/contracts/${id}`)
}
export function getLinggongContractByNo(contractNo) {
  return request.get(`/linggong/contracts/no/${contractNo}`)
}
export function createLinggongContract(data) {
  return request.post('/linggong/contracts', data)
}
export function updateLinggongContract(id, data) {
  return request.put(`/linggong/contracts/${id}`, data)
}
export function deleteLinggongContract(id) {
  return request.delete(`/linggong/contracts/${id}`)
}
export function signLinggongContract(id, data) {
  return request.post(`/linggong/contracts/${id}/sign`, data)
}
export function updateLinggongContractStatus(id, status) {
  return request.put(`/linggong/contracts/${id}/status`, { status })
}
export function listContractsByLinggong(linggongId, params) {
  return request.get(`/linggong/${linggongId}/contracts`, { params })
}
export function listContractsByEmployer(employerId, params) {
  return request.get(`/linggong/contracts/employer/${employerId}`, { params })
}
export function listContractsByWorker(workerId, params) {
  return request.get(`/linggong/contracts/worker/${workerId}`, { params })
}

// 管理端
export function adminListLinggongContracts(params) {
  return request.get('/linggong/admin/contracts', { params })
}

// ====================================================================
// 七、支付（linggong_payments）
// ====================================================================

export function getLinggongPayment(id) {
  return request.get(`/linggong/payments/${id}`)
}
export function getLinggongPaymentByNo(paymentNo) {
  return request.get(`/linggong/payments/no/${paymentNo}`)
}
export function createLinggongPayment(data) {
  return request.post('/linggong/payments', data)
}
export function updateLinggongPayment(id, data) {
  return request.put(`/linggong/payments/${id}`, data)
}
export function updateLinggongPaymentStatus(id, status) {
  return request.put(`/linggong/payments/${id}/status`, { status })
}
export function settleLinggongPayment(id, data) {
  return request.post(`/linggong/payments/${id}/settle`, data)
}
export function listPaymentsByLinggong(linggongId, params) {
  return request.get(`/linggong/${linggongId}/payments`, { params })
}
export function listPaymentsByEmployer(employerId, params) {
  return request.get(`/linggong/payments/employer/${employerId}`, { params })
}
export function listPaymentsByWorker(workerId, params) {
  return request.get(`/linggong/payments/worker/${workerId}`, { params })
}

// 管理端
export function adminListLinggongPayments(params) {
  return request.get('/linggong/admin/payments', { params })
}

// ====================================================================
// 八、评价（linggong_ratings）
// ====================================================================

export function listLinggongRatings(params) {
  return request.get('/linggong/ratings', { params })
}
export function getLinggongRating(id) {
  return request.get(`/linggong/ratings/${id}`)
}
export function getLinggongRatingByNo(ratingNo) {
  return request.get(`/linggong/ratings/no/${ratingNo}`)
}
export function getLinggongRatingStats(params) {
  return request.get('/linggong/ratings/stats', { params })
}
export function createLinggongRating(data) {
  return request.post('/linggong/ratings', data)
}
export function updateLinggongRating(id, data) {
  return request.put(`/linggong/ratings/${id}`, data)
}
export function deleteLinggongRating(id) {
  return request.delete(`/linggong/ratings/${id}`)
}
export function replyLinggongRating(id, data) {
  return request.post(`/linggong/ratings/${id}/reply`, data)
}
export function appendLinggongRating(id, data) {
  return request.post(`/linggong/ratings/${id}/append`, data)
}
export function likeLinggongRating(id) {
  return request.post(`/linggong/ratings/${id}/like`)
}
export function listRatingsByLinggong(linggongId, params) {
  return request.get(`/linggong/${linggongId}/ratings`, { params })
}
export function listRatingsByTarget(targetType, targetId, params) {
  return request.get(`/linggong/ratings/target/${targetType}/${targetId}`, { params })
}
export function listRatingsByRater(params) {
  return request.get('/linggong/ratings/rater', { params })
}

// 管理端
export function adminListLinggongRatings(params) {
  return request.get('/linggong/admin/ratings', { params })
}
export function auditLinggongRating(id, data) {
  return request.put(`/linggong/admin/ratings/${id}/audit`, data)
}

// ====================================================================
// 九、技能标签（linggong_skills）
// ====================================================================

export function listLinggongSkills(params) {
  return request.get('/linggong/skills', { params })
}
export function getLinggongSkill(id) {
  return request.get(`/linggong/skills/${id}`)
}
export function listHotLinggongSkills(params) {
  return request.get('/linggong/skills/hot', { params })
}
export function listSkillsByCategory(category, params) {
  return request.get(`/linggong/skills/category/${category}`, { params })
}
export function listSkillsByParent(parentId, params) {
  return request.get(`/linggong/skills/parent/${parentId}`, { params })
}

// 管理端 CRUD
export function adminListLinggongSkills(params) {
  return request.get('/linggong/admin/skills', { params })
}
export function createLinggongSkill(data) {
  return request.post('/linggong/skills', data)
}
export function updateLinggongSkill(id, data) {
  return request.put(`/linggong/skills/${id}`, data)
}
export function deleteLinggongSkill(id) {
  return request.delete(`/linggong/skills/${id}`)
}
export function adminUpdateLinggongSkillStatus(id, status) {
  return request.put(`/linggong/skills/${id}/status`, { status })
}

// ====================================================================
// 十、资质证书（linggong_certifications）
// ====================================================================

export function listLinggongCertifications(params) {
  return request.get('/linggong/certifications', { params })
}
export function getLinggongCertification(id) {
  return request.get(`/linggong/certifications/${id}`)
}
export function createLinggongCertification(data) {
  return request.post('/linggong/certifications', data)
}
export function updateLinggongCertification(id, data) {
  return request.put(`/linggong/certifications/${id}`, data)
}
export function deleteLinggongCertification(id) {
  return request.delete(`/linggong/certifications/${id}`)
}
export function listCertificationsByWorker(workerId, params) {
  return request.get(`/linggong/workers/${workerId}/certifications`, { params })
}
export function listMyLinggongCertifications(params) {
  return request.get('/linggong/certifications/mine', { params })
}

// 管理端
export function verifyLinggongCertification(id, data) {
  return request.put(`/linggong/admin/certifications/${id}/verify`, data)
}

// ====================================================================
// 十一、信用分（linggong_credits）
// ====================================================================

export function listLinggongCredits(params) {
  return request.get('/linggong/credits', { params })
}
export function getLinggongCredit(id) {
  return request.get(`/linggong/credits/${id}`)
}
export function listCreditsByUser(userId, params) {
  return request.get(`/linggong/credits/user/${userId}`, { params })
}
export function getMyLinggongCreditScore() {
  return request.get('/linggong/credits/score')
}

// 管理端
export function adjustLinggongCredit(data) {
  return request.post('/linggong/admin/credits/adjust', data)
}
export function deleteLinggongCredit(id) {
  return request.delete(`/linggong/admin/credits/${id}`)
}

// ====================================================================
// 十二、审核规则（linggong_audit_rules）
// ====================================================================

export function listLinggongAuditRules(params) {
  return request.get('/linggong/audit-rules', { params })
}
export function listEnabledLinggongAuditRules() {
  return request.get('/linggong/audit-rules/enabled')
}
export function listLinggongAuditRulesByType(ruleType, params) {
  return request.get(`/linggong/audit-rules/type/${ruleType}`, { params })
}
export function getLinggongAuditRule(id) {
  return request.get(`/linggong/audit-rules/${id}`)
}

// 管理端 CRUD
export function adminListLinggongAuditRules(params) {
  return request.get('/linggong/admin/audit-rules', { params })
}
export function createLinggongAuditRule(data) {
  return request.post('/linggong/admin/audit-rules', data)
}
export function updateLinggongAuditRule(id, data) {
  return request.put(`/linggong/admin/audit-rules/${id}`, data)
}
export function deleteLinggongAuditRule(id) {
  return request.delete(`/linggong/admin/audit-rules/${id}`)
}
export function adminUpdateLinggongAuditRuleStatus(id, status) {
  return request.put(`/linggong/admin/audit-rules/${id}/status`, { status })
}
export function batchDeleteLinggongAuditRules(data) {
  return request.post('/linggong/admin/audit-rules/batch-delete', data)
}

// ====================================================================
// 十三、统计（汇总统计 + 概览）
// ====================================================================

export function getLinggongOverviewStats(params) {
  return request.get('/linggong/admin/statistics/overview', { params }).catch(() => ({ data: {} }))
}
export function getLinggongHotSkills(params) {
  return request.get('/linggong/skills/hot', { params })
}
export function getLinggongRatingOverviewStats(params) {
  return request.get('/linggong/ratings/stats', { params })
}
