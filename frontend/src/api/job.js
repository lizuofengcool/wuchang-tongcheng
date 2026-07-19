// 同城招聘求职模块 API 封装（完整版 v1.0）
// 对应后端路由前缀：/api/v1/job（公开/C端） + /api/v1/job/admin（管理后台）
// 涵盖：职位/搜索/附近/收藏/公司/认证/简历/投递/面试/消息/会话/评价/举报/统计/审核
import request from '@/utils/request'

// ====================================================================
// 一、职位管理（公开 + C 端登录 + 管理后台）
// ====================================================================

// --- 公开接口（无需登录） ---

// 职位列表（C 端，已发布+已审核通过）
export function listJobs(params) {
  return request.get('/job', { params })
}

// 搜索职位
export function searchJobs(params) {
  return request.get('/job/search', { params })
}

// 高级搜索
export function advancedSearchJobs(params) {
  return request.get('/job/advanced-search', { params })
}

// 附近职位
export function listNearbyJobs(params) {
  return request.get('/job/nearby', { params })
}

// 职位详情
export function getJob(id) {
  return request.get(`/job/${id}`)
}

// 相似职位
export function listSimilarJobs(id, params) {
  return request.get(`/job/${id}/similar`, { params })
}

// 收藏状态（公开）
export function getJobFavStatus(id) {
  return request.get(`/job/${id}/fav`)
}

// 职位下的投递列表
export function listJobApplications(id, params) {
  return request.get(`/job/${id}/applications`, { params })
}

// --- C 端登录接口 ---

// 我的发布
export function listMyJobs(params) {
  return request.get('/job/mine', { params })
}

// 我的收藏
export function listMyJobFavs(params) {
  return request.get('/job/favorites', { params })
}

// 发布职位
export function createJob(data) {
  return request.post('/job', data)
}

// 更新职位
export function updateJob(id, data) {
  return request.put(`/job/${id}`, data)
}

// 删除职位
export function deleteJob(id) {
  return request.delete(`/job/${id}`)
}

// 上下架/状态变更
export function updateJobStatus(id, data) {
  return request.put(`/job/${id}/status`, data)
}

// 收藏 / 取消收藏（toggle 语义）
export function toggleJobFav(id) {
  return request.post(`/job/${id}/fav`)
}

// 取消收藏
export function unfavJob(id) {
  return request.delete(`/job/${id}/fav`)
}

// 推广职位
export function promoteJob(id, data) {
  return request.post(`/job/${id}/promotion`, data)
}

// --- 管理后台接口（需 job:audit 权限） ---

// 管理端列表（分页 + 筛选）
export function adminListJobs(params) {
  return request.get('/job/admin/list', { params })
}

// 管理端详情
export function adminGetJob(id) {
  return request.get(`/job/admin/${id}`)
}

// 审核（audit_status: 0待审 1通过 2拒绝；audit_reason 可选）
export function auditJob(id, data) {
  return request.put(`/job/admin/${id}/audit`, data)
}

// 强制上下架/状态变更（status: 1发布 3下架 4过期）
export function adminUpdateJobStatus(id, status) {
  return request.put(`/job/admin/${id}/status`, { status })
}

// ====================================================================
// 二、公司管理
// ====================================================================

// 公司列表（公开）
export function listCompanies(params) {
  return request.get('/job/companies', { params })
}

// 公司详情
export function getCompany(id) {
  return request.get(`/job/companies/${id}`)
}

// 公司认证列表（公开）
export function listCertifications(params) {
  return request.get('/job/certifications', { params })
}

// 公司认证详情
export function getCertification(id) {
  return request.get(`/job/certifications/${id}`)
}

// 公司下认证列表
export function listCompanyCertifications(companyId) {
  return request.get(`/job/companies/${companyId}/certifications`)
}

// 公司评价列表（公开）
export function listCompanyReviews(companyId, params) {
  return request.get(`/job/companies/${companyId}/reviews`, { params })
}

// 公司评价统计
export function getCompanyReviewStats(companyId) {
  return request.get(`/job/companies/${companyId}/reviews/stats`)
}

// 创建公司（C 端）
export function createCompany(data) {
  return request.post('/job/companies', data)
}

// 更新公司
export function updateCompany(id, data) {
  return request.put(`/job/companies/${id}`, data)
}

// 我的店铺
export function getMyCompany() {
  return request.get('/job/companies/mine')
}

// 关注公司
export function followCompany(id) {
  return request.post(`/job/companies/${id}/follow`, {})
}

// 取消关注
export function unfollowCompany(id) {
  return request.delete(`/job/companies/${id}/follow`)
}

// 我关注的公司
export function listFollowingCompanies(params) {
  return request.get('/job/companies/following', { params })
}

// 提交企业认证
export function createCertification(data) {
  return request.post('/job/certifications', data)
}

// 管理端 - 公司审核
export function auditCompany(id, data) {
  return request.put(`/job/admin/companies/${id}/audit`, data)
}

// 管理端 - 企业认证审核
export function processCertification(id, data) {
  return request.put(`/job/admin/certifications/${id}/process`, data)
}

// ====================================================================
// 三、简历管理
// ====================================================================

// 简历列表
export function listResumes(params) {
  return request.get('/job/resumes', { params })
}

// 我的简历
export function listMyResumes(params) {
  return request.get('/job/resumes/mine', { params })
}

// 默认简历
export function getDefaultResume() {
  return request.get('/job/resumes/default')
}

// 简历详情
export function getResume(id) {
  return request.get(`/job/resumes/${id}`)
}

// 创建简历
export function createResume(data) {
  return request.post('/job/resumes', data)
}

// 更新简历
export function updateResume(id, data) {
  return request.put(`/job/resumes/${id}`, data)
}

// 删除简历
export function deleteResume(id) {
  return request.delete(`/job/resumes/${id}`)
}

// 设为默认简历
export function setDefaultResume(id) {
  return request.put(`/job/resumes/${id}/default`)
}

// 简历状态变更
export function updateResumeStatus(id, data) {
  return request.put(`/job/resumes/${id}/status`, data)
}

// ====================================================================
// 四、投递记录
// ====================================================================

// 投递列表
export function listApplications(params) {
  return request.get('/job/applications', { params })
}

// 投递详情
export function getApplication(id) {
  return request.get(`/job/applications/${id}`)
}

// 投递简历
export function createApplication(data) {
  return request.post('/job/applications', data)
}

// 投递状态变更
export function updateApplicationStatus(id, data) {
  return request.put(`/job/applications/${id}/status`, data)
}

// 投递统计
export function getApplicationStats(params) {
  return request.get('/job/applications/stats', { params })
}

// 批量操作投递
export function batchActionApplications(data) {
  return request.post('/job/applications/batch', data)
}

// ====================================================================
// 五、面试邀约
// ====================================================================

// 面试列表
export function listInterviews(params) {
  return request.get('/job/interviews', { params })
}

// 面试详情
export function getInterview(id) {
  return request.get(`/job/interviews/${id}`)
}

// 创建面试邀约
export function createInterview(data) {
  return request.post('/job/interviews', data)
}

// 更新面试
export function updateInterview(id, data) {
  return request.put(`/job/interviews/${id}`, data)
}

// 面试动作（accept/reject/reschedule）
export function interviewAction(id, data) {
  return request.put(`/job/interviews/${id}/action`, data)
}

// 面试反馈
export function interviewFeedback(id, data) {
  return request.put(`/job/interviews/${id}/feedback`, data)
}

// 面试统计
export function getInterviewStats(params) {
  return request.get('/job/interviews/stats', { params })
}

// 投递下的面试列表
export function listInterviewsByApplication(applicationId, params) {
  return request.get(`/job/applications/${applicationId}/interviews`, { params })
}

// ====================================================================
// 六、沟通消息
// ====================================================================

// 消息列表
export function listMessages(params) {
  return request.get('/job/messages', { params })
}

// 未读消息数
export function countUnreadMessages() {
  return request.get('/job/messages/unread/count')
}

// 消息详情
export function getMessage(id) {
  return request.get(`/job/messages/${id}`)
}

// 发送消息
export function createMessage(data) {
  return request.post('/job/messages', data)
}

// 删除消息
export function deleteMessage(id) {
  return request.delete(`/job/messages/${id}`)
}

// 批量删除消息
export function batchDeleteMessages(data) {
  return request.post('/job/messages/batch-delete', data)
}

// 撤回消息
export function recallMessage(id) {
  return request.put(`/job/messages/${id}/recall`)
}

// 会话列表
export function listConversations(params) {
  return request.get('/job/conversations', { params })
}

// 会话消息列表
export function listConversationMessages(conversationId, params) {
  return request.get(`/job/conversations/${conversationId}/messages`, { params })
}

// 标记会话已读
export function markConversationRead(conversationId) {
  return request.put(`/job/conversations/${conversationId}/read`)
}

// ====================================================================
// 七、评价管理
// ====================================================================

// 评价列表（公开）
export function listReviews(params) {
  return request.get('/job/reviews', { params })
}

// 评价详情
export function getReview(id) {
  return request.get(`/job/reviews/${id}`)
}

// 用户评价列表
export function listUserReviews(userId, params) {
  return request.get(`/job/users/${userId}/reviews`, { params })
}

// 我的评价
export function listMyReviews(params) {
  return request.get('/job/reviews/mine', { params })
}

// 创建评价
export function createReview(data) {
  return request.post('/job/reviews', data)
}

// 更新评价
export function updateReview(id, data) {
  return request.put(`/job/reviews/${id}`, data)
}

// 删除评价
export function deleteReview(id) {
  return request.delete(`/job/reviews/${id}`)
}

// 评价回复
export function replyReview(id, data) {
  return request.post(`/job/reviews/${id}/reply`, data)
}

// 评价追加
export function appendReview(id, data) {
  return request.post(`/job/reviews/${id}/append`, data)
}

// 评价点赞
export function likeReview(id) {
  return request.post(`/job/reviews/${id}/like`)
}

// 管理端 - 评价审核
export function auditReview(id, data) {
  return request.put(`/job/admin/reviews/${id}/audit`, data)
}

// ====================================================================
// 八、举报管理
// ====================================================================

// 创建举报
export function createReport(data) {
  return request.post('/job/reports', data)
}

// 我的举报
export function listMyReports(params) {
  return request.get('/job/reports/mine', { params })
}

// 举报申诉
export function appealReport(id, data) {
  return request.post(`/job/reports/${id}/appeal`, data)
}

// 管理端 - 举报列表
export function adminListReports(params) {
  return request.get('/job/admin/reports', { params })
}

// 管理端 - 举报详情
export function adminGetReport(id) {
  return request.get(`/job/admin/reports/${id}`)
}

// 管理端 - 处理举报
export function processReport(id, data) {
  return request.put(`/job/admin/reports/${id}/process`, data)
}

// 管理端 - 申诉处理
export function processAppeal(id, data) {
  return request.put(`/job/admin/reports/${id}/appeal`, data)
}

// ====================================================================
// 九、数据统计
// ====================================================================

// 公开 - 热门职位
export function getHotJobs(params) {
  return request.get('/job/statistics/hot-jobs', { params })
}

// 公开 - 薪资趋势
export function getSalaryTrend(params) {
  return request.get('/job/statistics/salary-trend', { params })
}

// 公开 - 分类统计
export function getCategoryStats(params) {
  return request.get('/job/statistics/category', { params })
}

// 公开 - 分类列表（基于分类统计接口返回，job 模块无独立分类 CRUD）
export function listCategories(params) {
  return request.get('/job/statistics/category', { params })
}

// 公开 - 地区统计
export function getRegionStats(params) {
  return request.get('/job/statistics/region', { params })
}

// 公开 - 职位趋势
export function getJobTrendStats(params) {
  return request.get('/job/statistics/job-trend', { params })
}

// C 端 - 招聘者统计
export function getRecruiterStats(params) {
  return request.get('/job/statistics/recruiter', { params })
}

// C 端 - 求职者统计
export function getApplicantStats(params) {
  return request.get('/job/statistics/applicant', { params })
}

// C 端 - 仪表盘
export function getDashboardStats() {
  return request.get('/job/statistics/dashboard')
}

// 管理端 - 平台总览
export function getOverviewStats() {
  return request.get('/job/admin/statistics/overview')
}

// 管理端 - 转化漏斗
export function getConversionFunnel(params) {
  return request.get('/job/admin/statistics/conversion', { params })
}

// ====================================================================
// 兼容别名（便于其他模块统一引用）
// ====================================================================

export const adminListJobApplications = listJobApplications
export const getJobDetail = adminGetJob
export const getOverviewStatsJob = getOverviewStats
