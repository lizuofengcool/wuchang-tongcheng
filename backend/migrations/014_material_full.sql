-- ============================================================
-- 014_material_full.sql 物料/以图搜图中台扩展表
-- 对标淘宝拍立淘/京东拍照购/Google Lens
-- 创建 material_categories / material_tags / material_image_tags
-- / material_search_history / material_similar_results / material_ocr_results
-- 现有的 mat_files / mat_images / mat_videos / mat_image_features 已在 005 创建
-- 注：任务清单要求 material_images 表，实际由 mat_images 充当，本文件仅创建扩展表
-- 全部幂等：CREATE TABLE IF NOT EXISTS
-- ============================================================

-- ============================================================
-- 1. material_categories 图片分类
-- ============================================================
CREATE TABLE IF NOT EXISTS material_categories (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    name VARCHAR(64) NOT NULL,
    parent_id BIGINT NOT NULL DEFAULT 0,                         -- 父分类ID
    level SMALLINT NOT NULL DEFAULT 1,                            -- 层级
    icon VARCHAR(256) NOT NULL DEFAULT '',
    description VARCHAR(256) NOT NULL DEFAULT '',
    sort INTEGER NOT NULL DEFAULT 0,
    image_count INTEGER NOT NULL DEFAULT 0,                      -- 图片数量
    status SMALLINT NOT NULL DEFAULT 1,                          -- 1启用 0禁用
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_material_categories UNIQUE (name, parent_id)
);
CREATE INDEX IF NOT EXISTS idx_material_categories_parent_id ON material_categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_material_categories_status ON material_categories(status);
CREATE INDEX IF NOT EXISTS idx_material_categories_region_id ON material_categories(region_id);
COMMENT ON TABLE material_categories IS '物料图片分类';

-- ============================================================
-- 2. material_tags 图片标签
-- ============================================================
CREATE TABLE IF NOT EXISTS material_tags (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    name VARCHAR(64) NOT NULL,
    tag_type VARCHAR(32) NOT NULL DEFAULT 'general',              -- general/object/scene/color/shape
    description VARCHAR(256) NOT NULL DEFAULT '',
    icon VARCHAR(256) NOT NULL DEFAULT '',
    usage_count INTEGER NOT NULL DEFAULT 0,                      -- 使用次数
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_material_tags UNIQUE (name)
);
CREATE INDEX IF NOT EXISTS idx_material_tags_type ON material_tags(tag_type);
CREATE INDEX IF NOT EXISTS idx_material_tags_status ON material_tags(status);
CREATE INDEX IF NOT EXISTS idx_material_tags_usage ON material_tags(usage_count DESC);
CREATE INDEX IF NOT EXISTS idx_material_tags_region_id ON material_tags(region_id);
COMMENT ON TABLE material_tags IS '物料图片标签';

-- ============================================================
-- 3. material_image_tags 图片-标签关联表
-- ============================================================
CREATE TABLE IF NOT EXISTS material_image_tags (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    image_id BIGINT NOT NULL,                                    -- mat_images.id
    tag_id BIGINT NOT NULL,
    source VARCHAR(16) NOT NULL DEFAULT 'manual',                -- manual/auto（AI自动）
    confidence DECIMAL(5,2) NOT NULL DEFAULT 1.00,                -- 置信度
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_material_image_tags UNIQUE (image_id, tag_id)
);
CREATE INDEX IF NOT EXISTS idx_material_image_tags_tag_id ON material_image_tags(tag_id);
CREATE INDEX IF NOT EXISTS idx_material_image_tags_image_id ON material_image_tags(image_id);
CREATE INDEX IF NOT EXISTS idx_material_image_tags_region_id ON material_image_tags(region_id);
COMMENT ON TABLE material_image_tags IS '物料图片-标签关联';

-- ============================================================
-- 4. material_search_history 搜索历史
-- ============================================================
CREATE TABLE IF NOT EXISTS material_search_history (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    user_id BIGINT NOT NULL,
    search_type VARCHAR(16) NOT NULL DEFAULT 'image',            -- image/text/feature
    query_image_id BIGINT NOT NULL DEFAULT 0,                    -- 查询图片ID
    query_text VARCHAR(256) NOT NULL DEFAULT '',                  -- 查询文本
    result_count INTEGER NOT NULL DEFAULT 0,                     -- 返回结果数
    top_similarity DECIMAL(5,2) NOT NULL DEFAULT 0.00,            -- 最高相似度
    cost_ms INTEGER NOT NULL DEFAULT 0,                            -- 耗时（毫秒）
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_material_search_history_user_id ON material_search_history(user_id);
CREATE INDEX IF NOT EXISTS idx_material_search_history_created_at ON material_search_history(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_material_search_history_region_id ON material_search_history(region_id);
COMMENT ON TABLE material_search_history IS '物料搜索历史';

-- ============================================================
-- 5. material_similar_results 相似图搜索结果记录
-- ============================================================
CREATE TABLE IF NOT EXISTS material_similar_results (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    search_history_id BIGINT NOT NULL DEFAULT 0,
    source_image_id BIGINT NOT NULL,                             -- 源图片ID
    target_image_id BIGINT NOT NULL,                             -- 目标图片ID
    similarity DECIMAL(5,2) NOT NULL DEFAULT 0.00,               -- 相似度 0-1
    feature_algo VARCHAR(32) NOT NULL DEFAULT 'phash',            -- 算法 phash/cnn/color_hist
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_material_similar_source ON material_similar_results(source_image_id);
CREATE INDEX IF NOT EXISTS idx_material_similar_target ON material_similar_results(target_image_id);
CREATE INDEX IF NOT EXISTS idx_material_similar_algo ON material_similar_results(feature_algo);
CREATE INDEX IF NOT EXISTS idx_material_similar_region_id ON material_similar_results(region_id);
COMMENT ON TABLE material_similar_results IS '物料相似图搜索结果';

-- ============================================================
-- 6. material_ocr_results OCR 识别结果
-- ============================================================
CREATE TABLE IF NOT EXISTS material_ocr_results (
    id BIGSERIAL PRIMARY KEY,
    region_id BIGINT NOT NULL DEFAULT 1,
    image_id BIGINT NOT NULL,                                    -- mat_images.id
    file_id VARCHAR(64) NOT NULL DEFAULT '',
    ocr_engine VARCHAR(32) NOT NULL DEFAULT 'local',             -- 引擎 local/aliyun/tencent/baidu
    text_content TEXT NOT NULL DEFAULT '',                       -- 识别出的文本
    text_blocks JSONB NOT NULL DEFAULT '[]'::jsonb,               -- 文本块结构
    language VARCHAR(16) NOT NULL DEFAULT 'zh',                  -- 识别语言
    confidence DECIMAL(5,2) NOT NULL DEFAULT 0.00,                -- 整体置信度
    cost_ms INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_material_ocr_image_id ON material_ocr_results(image_id);
CREATE INDEX IF NOT EXISTS idx_material_ocr_file_id ON material_ocr_results(file_id);
CREATE INDEX IF NOT EXISTS idx_material_ocr_engine ON material_ocr_results(ocr_engine);
CREATE INDEX IF NOT EXISTS idx_material_ocr_region_id ON material_ocr_results(region_id);
COMMENT ON TABLE material_ocr_results IS '物料图片 OCR 识别结果';
