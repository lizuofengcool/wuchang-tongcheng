-- ============================================================
-- pinche 拼车出行模块完整功能迁移脚本（v3.2.1）
-- 对标：哈啰出行 / 嘀嗒出行 / 滴滴顺风车
--
-- 内容：
--   1. CREATE 18 张表（pinches 主表 + 17 张子表，pinche_ 前缀）
--   2. 索引、触发器、注释
--   3. 全幂等：CREATE TABLE IF NOT EXISTS
-- 依赖：001_p0_baseline.sql（update_updated_at_column 函数已创建）
--
-- 注意：pinches 主表也会在 plugin.go Init 中由 GORM AutoMigrate 创建
--      此处 CREATE TABLE IF NOT EXISTS 保证幂等，二者不冲突
-- ============================================================

-- ============================================================
-- 1. pinches 拼车行程主表（也由 GORM AutoMigrate 创建）
-- ============================================================
CREATE TABLE IF NOT EXISTS pinches (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 发布者信息
    user_id BIGINT NOT NULL,
    user_name VARCHAR(50) NOT NULL DEFAULT '',
    user_phone VARCHAR(20) NOT NULL DEFAULT '',
    user_avatar VARCHAR(255) NOT NULL DEFAULT '',

    -- 行程类型与角色
    trip_type VARCHAR(16) NOT NULL DEFAULT 'shunfeng',  -- shunfeng顺风车/pinche拼车/baoche包车
    role VARCHAR(16) NOT NULL DEFAULT 'driver',          -- driver车主/passenger乘客

    -- 基础信息
    title VARCHAR(200) NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    cover_image VARCHAR(255) NOT NULL DEFAULT '',

    -- 状态
    status INT NOT NULL DEFAULT 0,           -- 0草稿 1已发布 2已结束 3已取消 4进行中
    audit_status INT NOT NULL DEFAULT 0,     -- 0待审 1通过 2拒绝
    audit_reason VARCHAR(500) NOT NULL DEFAULT '',
    published_at TIMESTAMPTZ,

    -- 出发信息
    departure_time TIMESTAMPTZ,
    pickup_location VARCHAR(255) NOT NULL DEFAULT '',
    pickup_lat DECIMAL(10,7) NOT NULL DEFAULT 0,
    pickup_lng DECIMAL(10,7) NOT NULL DEFAULT 0,
    dropoff_location VARCHAR(255) NOT NULL DEFAULT '',
    dropoff_lat DECIMAL(10,7) NOT NULL DEFAULT 0,
    dropoff_lng DECIMAL(10,7) NOT NULL DEFAULT 0,

    -- 行程距离/时长
    distance_km DECIMAL(10,2) NOT NULL DEFAULT 0,
    duration_min INT NOT NULL DEFAULT 0,

    -- 座位
    total_seats INT NOT NULL DEFAULT 4,
    available_seats INT NOT NULL DEFAULT 4,
    booked_seats INT NOT NULL DEFAULT 0,

    -- 金额
    price_per_seat DECIMAL(12,2) NOT NULL DEFAULT 0,
    total_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    toll_fee DECIMAL(12,2) NOT NULL DEFAULT 0,

    -- 关联
    vehicle_id BIGINT,
    driver_id BIGINT,
    route_id BIGINT,
    insurance_id BIGINT,
    trip_id BIGINT,
    emergency_contact_id BIGINT,

    -- 行程分享与支付
    share_token VARCHAR(64) NOT NULL DEFAULT '',
    payment_method VARCHAR(16) NOT NULL DEFAULT 'cash', -- cash/wechat/alipay/balance/etc

    -- 配置/特征（JSONB）
    features JSONB,
    tags JSONB,

    -- 互动统计
    view_count INT NOT NULL DEFAULT 0,
    fav_count INT NOT NULL DEFAULT 0,
    contact_count INT NOT NULL DEFAULT 0,
    share_count INT NOT NULL DEFAULT 0,

    -- 运营字段
    featured BOOLEAN NOT NULL DEFAULT FALSE,
    picked BOOLEAN NOT NULL DEFAULT FALSE,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    promotion_level INT NOT NULL DEFAULT 0,

    -- 风控
    content_hash VARCHAR(64) NOT NULL DEFAULT '',
    risk_score INT NOT NULL DEFAULT 0,

    -- 时间节点
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_pinches_region_id ON pinches(region_id);
CREATE INDEX IF NOT EXISTS idx_pinches_user_id ON pinches(user_id);
CREATE INDEX IF NOT EXISTS idx_pinches_status ON pinches(status);
CREATE INDEX IF NOT EXISTS idx_pinches_audit_status ON pinches(audit_status);
CREATE INDEX IF NOT EXISTS idx_pinches_trip_type ON pinches(trip_type);
CREATE INDEX IF NOT EXISTS idx_pinches_role ON pinches(role);
CREATE INDEX IF NOT EXISTS idx_pinches_departure_time ON pinches(departure_time);
CREATE INDEX IF NOT EXISTS idx_pinches_driver_id ON pinches(driver_id);
CREATE INDEX IF NOT EXISTS idx_pinches_route_id ON pinches(route_id);
CREATE INDEX IF NOT EXISTS idx_pinches_deleted_at ON pinches(deleted_at);
CREATE INDEX IF NOT EXISTS idx_pinches_featured ON pinches(featured) WHERE featured = TRUE;
CREATE INDEX IF NOT EXISTS idx_pinches_published_at ON pinches(published_at);

COMMENT ON TABLE pinches IS '拼车行程主表';
COMMENT ON COLUMN pinches.trip_type IS '行程类型：shunfeng顺风车/pinche拼车/baoche包车';
COMMENT ON COLUMN pinches.role IS '发布者身份：driver车主/passenger乘客';
COMMENT ON COLUMN pinches.status IS '状态：0草稿 1已发布 2已结束 3已取消 4进行中';
COMMENT ON COLUMN pinches.payment_method IS '支付方式：cash/wechat/alipay/balance/etc';

-- ============================================================
-- 2. pinche_routes 路线表
-- ============================================================
CREATE TABLE IF NOT EXISTS pinche_routes (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    user_id BIGINT NOT NULL,
    route_name VARCHAR(128) NOT NULL DEFAULT '',

    -- 起终点
    origin_address VARCHAR(255) NOT NULL DEFAULT '',
    origin_lat DECIMAL(10,7) NOT NULL DEFAULT 0,
    origin_lng DECIMAL(10,7) NOT NULL DEFAULT 0,
    destination_address VARCHAR(255) NOT NULL DEFAULT '',
    destination_lat DECIMAL(10,7) NOT NULL DEFAULT 0,
    destination_lng DECIMAL(10,7) NOT NULL DEFAULT 0,

    -- 途经点（JSONB 数组：[{address,lat,lng}]）
    waypoints JSONB,

    -- 行程信息
    distance_km DECIMAL(10,2) NOT NULL DEFAULT 0,
    duration_min INT NOT NULL DEFAULT 0,
    estimated_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    toll_fee DECIMAL(12,2) NOT NULL DEFAULT 0,

    is_common BOOLEAN NOT NULL DEFAULT FALSE,
    use_count INT NOT NULL DEFAULT 0,
    status INT NOT NULL DEFAULT 1
);
CREATE INDEX IF NOT EXISTS idx_pinche_routes_region_id ON pinche_routes(region_id);
CREATE INDEX IF NOT EXISTS idx_pinche_routes_user_id ON pinche_routes(user_id);
CREATE INDEX IF NOT EXISTS idx_pinche_routes_is_common ON pinche_routes(is_common);
CREATE INDEX IF NOT EXISTS idx_pinche_routes_deleted_at ON pinche_routes(deleted_at);

COMMENT ON TABLE pinche_routes IS '拼车路线表（起点/终点/途经点/距离/时长）';

-- ============================================================
-- 3. pinche_bookings 预订记录
-- ============================================================
CREATE TABLE IF NOT EXISTS pinche_bookings (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    pinche_id BIGINT NOT NULL,
    booking_no VARCHAR(32) NOT NULL,

    -- 乘客
    passenger_id BIGINT NOT NULL,
    passenger_name VARCHAR(50) NOT NULL DEFAULT '',
    passenger_phone VARCHAR(20) NOT NULL DEFAULT '',
    passenger_avatar VARCHAR(255) NOT NULL DEFAULT '',

    -- 车主
    driver_id BIGINT NOT NULL,
    driver_name VARCHAR(50) NOT NULL DEFAULT '',
    driver_phone VARCHAR(20) NOT NULL DEFAULT '',

    -- 预订信息
    seats INT NOT NULL DEFAULT 1,
    pickup_location VARCHAR(255) NOT NULL DEFAULT '',
    pickup_lat DECIMAL(10,7) NOT NULL DEFAULT 0,
    pickup_lng DECIMAL(10,7) NOT NULL DEFAULT 0,
    dropoff_location VARCHAR(255) NOT NULL DEFAULT '',
    dropoff_lat DECIMAL(10,7) NOT NULL DEFAULT 0,
    dropoff_lng DECIMAL(10,7) NOT NULL DEFAULT 0,

    -- 金额
    unit_price DECIMAL(12,2) NOT NULL DEFAULT 0,
    total_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    insurance_fee DECIMAL(12,2) NOT NULL DEFAULT 0,
    service_fee DECIMAL(12,2) NOT NULL DEFAULT 0,

    -- 状态
    status INT NOT NULL DEFAULT 0, -- 0待支付 1已支付 2已上车 3已完成 4已取消 5已退款
    payment_id BIGINT,
    boarding_code VARCHAR(16) NOT NULL DEFAULT '',

    -- 时间节点
    paid_at TIMESTAMPTZ,
    boarded_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    cancel_reason VARCHAR(500) NOT NULL DEFAULT '',
    cancelled_by BIGINT
);
CREATE INDEX IF NOT EXISTS idx_pinche_bookings_region_id ON pinche_bookings(region_id);
CREATE INDEX IF NOT EXISTS idx_pinche_bookings_pinche_id ON pinche_bookings(pinche_id);
CREATE INDEX IF NOT EXISTS idx_pinche_bookings_booking_no ON pinche_bookings(booking_no);
CREATE INDEX IF NOT EXISTS idx_pinche_bookings_passenger_id ON pinche_bookings(passenger_id);
CREATE INDEX IF NOT EXISTS idx_pinche_bookings_driver_id ON pinche_bookings(driver_id);
CREATE INDEX IF NOT EXISTS idx_pinche_bookings_status ON pinche_bookings(status);
CREATE INDEX IF NOT EXISTS idx_pinche_bookings_payment_id ON pinche_bookings(payment_id);
CREATE INDEX IF NOT EXISTS idx_pinche_bookings_deleted_at ON pinche_bookings(deleted_at);

COMMENT ON TABLE pinche_bookings IS '拼车预订记录表';

-- ============================================================
-- 4. pinche_drivers 车主认证表
-- ============================================================
CREATE TABLE IF NOT EXISTS pinche_drivers (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    user_id BIGINT NOT NULL,
    user_name VARCHAR(50) NOT NULL DEFAULT '',
    user_phone VARCHAR(20) NOT NULL DEFAULT '',
    user_avatar VARCHAR(255) NOT NULL DEFAULT '',

    -- 实名信息
    real_name VARCHAR(50) NOT NULL DEFAULT '',
    id_card_no VARCHAR(32) NOT NULL DEFAULT '',
    id_card_front VARCHAR(255) NOT NULL DEFAULT '',
    id_card_back VARCHAR(255) NOT NULL DEFAULT '',

    -- 驾驶证
    driver_license_no VARCHAR(32) NOT NULL DEFAULT '',
    driver_license_front VARCHAR(255) NOT NULL DEFAULT '',
    driver_license_back VARCHAR(255) NOT NULL DEFAULT '',
    license_issue_date DATE,
    license_expiry_date DATE,

    -- 行驶证与车辆
    vehicle_license_no VARCHAR(32) NOT NULL DEFAULT '',
    vehicle_license_front VARCHAR(255) NOT NULL DEFAULT '',
    vehicle_license_back VARCHAR(255) NOT NULL DEFAULT '',
    car_photo VARCHAR(255) NOT NULL DEFAULT '',

    -- 状态
    status INT NOT NULL DEFAULT 0, -- 0待审 1通过 2拒绝 3已过期
    audit_reason VARCHAR(500) NOT NULL DEFAULT '',
    audited_at TIMESTAMPTZ,
    auditor_id BIGINT,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at TIMESTAMPTZ,

    -- 统计
    rating_avg DECIMAL(3,2) NOT NULL DEFAULT 5.00,
    trip_count INT NOT NULL DEFAULT 0,
    total_income DECIMAL(12,2) NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_pinche_drivers_region_id ON pinche_drivers(region_id);
CREATE INDEX IF NOT EXISTS idx_pinche_drivers_user_id ON pinche_drivers(user_id);
CREATE INDEX IF NOT EXISTS idx_pinche_drivers_status ON pinche_drivers(status);
CREATE INDEX IF NOT EXISTS idx_pinche_drivers_verified ON pinche_drivers(verified) WHERE verified = TRUE;
CREATE INDEX IF NOT EXISTS idx_pinche_drivers_deleted_at ON pinche_drivers(deleted_at);

COMMENT ON TABLE pinche_drivers IS '拼车车主认证表（身份证+驾驶证+行驶证+车辆照片）';

-- ============================================================
-- 5. pinche_vehicles 车辆信息表
-- ============================================================
CREATE TABLE IF NOT EXISTS pinche_vehicles (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    driver_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,

    -- 车辆信息
    plate_no VARCHAR(16) NOT NULL DEFAULT '',
    brand VARCHAR(64) NOT NULL DEFAULT '',
    model VARCHAR(128) NOT NULL DEFAULT '',
    year INT NOT NULL DEFAULT 0,
    color VARCHAR(32) NOT NULL DEFAULT '',
    seat_count INT NOT NULL DEFAULT 5,
    vehicle_type VARCHAR(32) NOT NULL DEFAULT 'sedan', -- sedan/suv/mpv/new_energy
    fuel_type VARCHAR(32) NOT NULL DEFAULT 'gasoline', -- gasoline/electric/hybrid

    -- 照片
    vehicle_photos JSONB,                  -- 车辆照片数组
    vehicle_license_photo VARCHAR(255) NOT NULL DEFAULT '',
    insurance_photo VARCHAR(255) NOT NULL DEFAULT '',

    -- 状态
    status INT NOT NULL DEFAULT 0,          -- 0待审 1通过 2拒绝
    audit_status INT NOT NULL DEFAULT 0,
    audit_reason VARCHAR(500) NOT NULL DEFAULT '',
    is_default BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_pinche_vehicles_region_id ON pinche_vehicles(region_id);
CREATE INDEX IF NOT EXISTS idx_pinche_vehicles_driver_id ON pinche_vehicles(driver_id);
CREATE INDEX IF NOT EXISTS idx_pinche_vehicles_user_id ON pinche_vehicles(user_id);
CREATE INDEX IF NOT EXISTS idx_pinche_vehicles_plate_no ON pinche_vehicles(plate_no);
CREATE INDEX IF NOT EXISTS idx_pinche_vehicles_status ON pinche_vehicles(status);
CREATE INDEX IF NOT EXISTS idx_pinche_vehicles_deleted_at ON pinche_vehicles(deleted_at);

COMMENT ON TABLE pinche_vehicles IS '拼车车辆信息表';

-- ============================================================
-- 6. pinche_insurances 顺风车保险表
-- ============================================================
CREATE TABLE IF NOT EXISTS pinche_insurances (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    pinche_id BIGINT NOT NULL,
    booking_id BIGINT,

    policy_no VARCHAR(64) NOT NULL DEFAULT '',
    insurance_company VARCHAR(64) NOT NULL DEFAULT '',
    insurance_type VARCHAR(32) NOT NULL DEFAULT 'passenger', -- passenger/driver/both
    coverage_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    premium DECIMAL(12,2) NOT NULL DEFAULT 0,

    insured_name VARCHAR(50) NOT NULL DEFAULT '',
    insured_id_card VARCHAR(32) NOT NULL DEFAULT '',

    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,

    status INT NOT NULL DEFAULT 0, -- 0待生效 1生效中 2已结束 3已理赔
    claim_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    claim_reason VARCHAR(500) NOT NULL DEFAULT '',
    claimed_at TIMESTAMPTZ,

    beneficiaries JSONB
);
CREATE INDEX IF NOT EXISTS idx_pinche_insurances_region_id ON pinche_insurances(region_id);
CREATE INDEX IF NOT EXISTS idx_pinche_insurances_pinche_id ON pinche_insurances(pinche_id);
CREATE INDEX IF NOT EXISTS idx_pinche_insurances_booking_id ON pinche_insurances(booking_id);
CREATE INDEX IF NOT EXISTS idx_pinche_insurances_policy_no ON pinche_insurances(policy_no);
CREATE INDEX IF NOT EXISTS idx_pinche_insurances_status ON pinche_insurances(status);
CREATE INDEX IF NOT EXISTS idx_pinche_insurances_deleted_at ON pinche_insurances(deleted_at);

COMMENT ON TABLE pinche_insurances IS '拼车顺风车保险表';

-- ============================================================
-- 7. pinche_payments 支付记录表（含 ETC 支付）
-- ============================================================
CREATE TABLE IF NOT EXISTS pinche_payments (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    pinche_id BIGINT NOT NULL,
    booking_id BIGINT NOT NULL,

    payment_no VARCHAR(32) NOT NULL,
    payer_id BIGINT NOT NULL,
    payer_name VARCHAR(50) NOT NULL DEFAULT '',
    payee_id BIGINT NOT NULL,
    payee_name VARCHAR(50) NOT NULL DEFAULT '',

    amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    insurance_fee DECIMAL(12,2) NOT NULL DEFAULT 0,
    service_fee DECIMAL(12,2) NOT NULL DEFAULT 0,
    total_amount DECIMAL(12,2) NOT NULL DEFAULT 0,

    payment_method VARCHAR(16) NOT NULL DEFAULT 'cash', -- cash/wechat/alipay/balance/etc
    payment_status INT NOT NULL DEFAULT 0,             -- 0待支付 1已支付 2已退款 3已失败
    paid_at TIMESTAMPTZ,
    refunded_at TIMESTAMPTZ,
    third_party_no VARCHAR(64) NOT NULL DEFAULT '',
    refund_amount DECIMAL(12,2) NOT NULL DEFAULT 0,

    -- ETC 专用
    etc_lane_id VARCHAR(32) NOT NULL DEFAULT '',
    etc_entry_time TIMESTAMPTZ,
    etc_exit_time TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_pinche_payments_region_id ON pinche_payments(region_id);
CREATE INDEX IF NOT EXISTS idx_pinche_payments_pinche_id ON pinche_payments(pinche_id);
CREATE INDEX IF NOT EXISTS idx_pinche_payments_booking_id ON pinche_payments(booking_id);
CREATE INDEX IF NOT EXISTS idx_pinche_payments_payment_no ON pinche_payments(payment_no);
CREATE INDEX IF NOT EXISTS idx_pinche_payments_payer_id ON pinche_payments(payer_id);
CREATE INDEX IF NOT EXISTS idx_pinche_payments_payee_id ON pinche_payments(payee_id);
CREATE INDEX IF NOT EXISTS idx_pinche_payments_status ON pinche_payments(payment_status);
CREATE INDEX IF NOT EXISTS idx_pinche_payments_method ON pinche_payments(payment_method);
CREATE INDEX IF NOT EXISTS idx_pinche_payments_deleted_at ON pinche_payments(deleted_at);

COMMENT ON TABLE pinche_payments IS '拼车支付记录表（含 ETC 支付）';

-- ============================================================
-- 8. pinche_ratings 评价表
-- ============================================================
CREATE TABLE IF NOT EXISTS pinche_ratings (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    pinche_id BIGINT NOT NULL,
    booking_id BIGINT,
    trip_id BIGINT,

    rater_id BIGINT NOT NULL,
    rater_name VARCHAR(50) NOT NULL DEFAULT '',
    rater_avatar VARCHAR(255) NOT NULL DEFAULT '',
    ratee_id BIGINT NOT NULL,
    ratee_name VARCHAR(50) NOT NULL DEFAULT '',
    ratee_avatar VARCHAR(255) NOT NULL DEFAULT '',

    rating_type VARCHAR(32) NOT NULL DEFAULT 'passenger_to_driver', -- passenger_to_driver/driver_to_passenger
    rating INT NOT NULL DEFAULT 5,
    content TEXT NOT NULL DEFAULT '',
    images JSONB,
    tags JSONB,
    is_anonymous BOOLEAN NOT NULL DEFAULT FALSE,

    reply TEXT NOT NULL DEFAULT '',
    reply_at TIMESTAMPTZ,
    like_count INT NOT NULL DEFAULT 0,

    status INT NOT NULL DEFAULT 0 -- 0待审 1通过 2拒绝 3隐藏
);
CREATE INDEX IF NOT EXISTS idx_pinche_ratings_region_id ON pinche_ratings(region_id);
CREATE INDEX IF NOT EXISTS idx_pinche_ratings_pinche_id ON pinche_ratings(pinche_id);
CREATE INDEX IF NOT EXISTS idx_pinche_ratings_booking_id ON pinche_ratings(booking_id);
CREATE INDEX IF NOT EXISTS idx_pinche_ratings_rater_id ON pinche_ratings(rater_id);
CREATE INDEX IF NOT EXISTS idx_pinche_ratings_ratee_id ON pinche_ratings(ratee_id);
CREATE INDEX IF NOT EXISTS idx_pinche_ratings_rating_type ON pinche_ratings(rating_type);
CREATE INDEX IF NOT EXISTS idx_pinche_ratings_status ON pinche_ratings(status);
CREATE INDEX IF NOT EXISTS idx_pinche_ratings_deleted_at ON pinche_ratings(deleted_at);

COMMENT ON TABLE pinche_ratings IS '拼车评价表';

-- ============================================================
-- 9. pinche_emergencies 紧急联系人/一键报警表
-- ============================================================
CREATE TABLE IF NOT EXISTS pinche_emergencies (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    -- 紧急联系人
    user_id BIGINT NOT NULL,
    contact_name VARCHAR(50) NOT NULL DEFAULT '',
    contact_phone VARCHAR(20) NOT NULL DEFAULT '',
    contact_relation VARCHAR(32) NOT NULL DEFAULT '',
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,

    -- 关联行程
    pinche_id BIGINT,
    trip_id BIGINT,

    -- 报警信息
    alert_type VARCHAR(16) NOT NULL DEFAULT 'sos', -- sos/share/periodic
    alert_status INT NOT NULL DEFAULT 0,            -- 0未处理 1处理中 2已处理
    alert_time TIMESTAMPTZ,
    alert_location_lat DECIMAL(10,7) NOT NULL DEFAULT 0,
    alert_location_lng DECIMAL(10,7) NOT NULL DEFAULT 0,
    alert_address VARCHAR(255) NOT NULL DEFAULT '',
    alert_description VARCHAR(500) NOT NULL DEFAULT '',
    alert_evidence JSONB,                           -- 录音/视频/照片 URL 数组

    handled_at TIMESTAMPTZ,
    handler_id BIGINT,
    handle_result VARCHAR(500) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_pinche_emergencies_region_id ON pinche_emergencies(region_id);
CREATE INDEX IF NOT EXISTS idx_pinche_emergencies_user_id ON pinche_emergencies(user_id);
CREATE INDEX IF NOT EXISTS idx_pinche_emergencies_pinche_id ON pinche_emergencies(pinche_id);
CREATE INDEX IF NOT EXISTS idx_pinche_emergencies_trip_id ON pinche_emergencies(trip_id);
CREATE INDEX IF NOT EXISTS idx_pinche_emergencies_alert_type ON pinche_emergencies(alert_type);
CREATE INDEX IF NOT EXISTS idx_pinche_emergencies_alert_status ON pinche_emergencies(alert_status);
CREATE INDEX IF NOT EXISTS idx_pinche_emergencies_deleted_at ON pinche_emergencies(deleted_at);

COMMENT ON TABLE pinche_emergencies IS '拼车紧急联系人/一键报警表';

-- ============================================================
-- 10. pinche_trips 完成行程表（含行程分享 share_token）
-- ============================================================
CREATE TABLE IF NOT EXISTS pinche_trips (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    pinche_id BIGINT NOT NULL,
    booking_id BIGINT,
    trip_no VARCHAR(32) NOT NULL,

    driver_id BIGINT NOT NULL,
    driver_name VARCHAR(50) NOT NULL DEFAULT '',
    driver_phone VARCHAR(20) NOT NULL DEFAULT '',
    passenger_id BIGINT NOT NULL,
    passenger_name VARCHAR(50) NOT NULL DEFAULT '',
    passenger_phone VARCHAR(20) NOT NULL DEFAULT '',
    vehicle_id BIGINT,
    plate_no VARCHAR(16) NOT NULL DEFAULT '',

    -- 行程信息
    origin_address VARCHAR(255) NOT NULL DEFAULT '',
    origin_lat DECIMAL(10,7) NOT NULL DEFAULT 0,
    origin_lng DECIMAL(10,7) NOT NULL DEFAULT 0,
    destination_address VARCHAR(255) NOT NULL DEFAULT '',
    destination_lat DECIMAL(10,7) NOT NULL DEFAULT 0,
    destination_lng DECIMAL(10,7) NOT NULL DEFAULT 0,

    -- 实际行程
    actual_pickup_time TIMESTAMPTZ,
    actual_dropoff_time TIMESTAMPTZ,
    actual_distance_km DECIMAL(10,2) NOT NULL DEFAULT 0,
    actual_duration_min INT NOT NULL DEFAULT 0,
    passengers_count INT NOT NULL DEFAULT 1,

    -- 金额
    fare_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    toll_fee DECIMAL(12,2) NOT NULL DEFAULT 0,
    total_amount DECIMAL(12,2) NOT NULL DEFAULT 0,

    -- 行程分享
    share_token VARCHAR(64) NOT NULL DEFAULT '',
    share_expires_at TIMESTAMPTZ,

    -- 状态
    status INT NOT NULL DEFAULT 0, -- 0进行中 1已完成 2异常结束
    driver_confirmed_at TIMESTAMPTZ,
    passenger_confirmed_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_pinche_trips_region_id ON pinche_trips(region_id);
CREATE INDEX IF NOT EXISTS idx_pinche_trips_pinche_id ON pinche_trips(pinche_id);
CREATE INDEX IF NOT EXISTS idx_pinche_trips_booking_id ON pinche_trips(booking_id);
CREATE INDEX IF NOT EXISTS idx_pinche_trips_trip_no ON pinche_trips(trip_no);
CREATE INDEX IF NOT EXISTS idx_pinche_trips_driver_id ON pinche_trips(driver_id);
CREATE INDEX IF NOT EXISTS idx_pinche_trips_passenger_id ON pinche_trips(passenger_id);
CREATE INDEX IF NOT EXISTS idx_pinche_trips_share_token ON pinche_trips(share_token);
CREATE INDEX IF NOT EXISTS idx_pinche_trips_status ON pinche_trips(status);
CREATE INDEX IF NOT EXISTS idx_pinche_trips_deleted_at ON pinche_trips(deleted_at);

COMMENT ON TABLE pinche_trips IS '拼车完成行程表（含行程分享 share_token）';

-- ============================================================
-- 11. pinche_route_favorites 常用路线收藏表
-- ============================================================
CREATE TABLE IF NOT EXISTS pinche_route_favorites (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    user_id BIGINT NOT NULL,
    route_id BIGINT,
    favorite_name VARCHAR(128) NOT NULL DEFAULT '',

    origin_address VARCHAR(255) NOT NULL DEFAULT '',
    origin_lat DECIMAL(10,7) NOT NULL DEFAULT 0,
    origin_lng DECIMAL(10,7) NOT NULL DEFAULT 0,
    destination_address VARCHAR(255) NOT NULL DEFAULT '',
    destination_lat DECIMAL(10,7) NOT NULL DEFAULT 0,
    destination_lng DECIMAL(10,7) NOT NULL DEFAULT 0,

    use_count INT NOT NULL DEFAULT 0,
    last_used_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_pinche_route_favorites_region_id ON pinche_route_favorites(region_id);
CREATE INDEX IF NOT EXISTS idx_pinche_route_favorites_user_id ON pinche_route_favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_pinche_route_favorites_route_id ON pinche_route_favorites(route_id);
CREATE INDEX IF NOT EXISTS idx_pinche_route_favorites_deleted_at ON pinche_route_favorites(deleted_at);

COMMENT ON TABLE pinche_route_favorites IS '拼车常用路线收藏表';

-- ============================================================
-- 12. pinche_driver_locations 实时位置表
-- ============================================================
CREATE TABLE IF NOT EXISTS pinche_driver_locations (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    pinche_id BIGINT NOT NULL,
    trip_id BIGINT,
    booking_id BIGINT,
    driver_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,

    lat DECIMAL(10,7) NOT NULL DEFAULT 0,
    lng DECIMAL(10,7) NOT NULL DEFAULT 0,
    speed DECIMAL(6,2) NOT NULL DEFAULT 0,
    heading DECIMAL(5,2) NOT NULL DEFAULT 0,
    altitude DECIMAL(8,2) NOT NULL DEFAULT 0,
    location_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    is_shared BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_pinche_driver_locations_pinche_id ON pinche_driver_locations(pinche_id);
CREATE INDEX IF NOT EXISTS idx_pinche_driver_locations_trip_id ON pinche_driver_locations(trip_id);
CREATE INDEX IF NOT EXISTS idx_pinche_driver_locations_driver_id ON pinche_driver_locations(driver_id);
CREATE INDEX IF NOT EXISTS idx_pinche_driver_locations_location_time ON pinche_driver_locations(location_time);

COMMENT ON TABLE pinche_driver_locations IS '拼车车主实时位置表';

-- ============================================================
-- 13. pinche_messages 行程内消息表
-- ============================================================
CREATE TABLE IF NOT EXISTS pinche_messages (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    pinche_id BIGINT NOT NULL,
    booking_id BIGINT,
    trip_id BIGINT,

    sender_id BIGINT NOT NULL,
    sender_name VARCHAR(50) NOT NULL DEFAULT '',
    sender_avatar VARCHAR(255) NOT NULL DEFAULT '',
    receiver_id BIGINT NOT NULL,
    receiver_name VARCHAR(50) NOT NULL DEFAULT '',

    message_type VARCHAR(16) NOT NULL DEFAULT 'text', -- text/image/voice/system/location
    content TEXT NOT NULL DEFAULT '',
    image_url VARCHAR(255) NOT NULL DEFAULT '',
    voice_url VARCHAR(255) NOT NULL DEFAULT '',
    voice_duration INT NOT NULL DEFAULT 0,
    location_lat DECIMAL(10,7) NOT NULL DEFAULT 0,
    location_lng DECIMAL(10,7) NOT NULL DEFAULT 0,
    location_address VARCHAR(255) NOT NULL DEFAULT '',

    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMPTZ,
    system_event VARCHAR(32) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_pinche_messages_region_id ON pinche_messages(region_id);
CREATE INDEX IF NOT EXISTS idx_pinche_messages_pinche_id ON pinche_messages(pinche_id);
CREATE INDEX IF NOT EXISTS idx_pinche_messages_booking_id ON pinche_messages(booking_id);
CREATE INDEX IF NOT EXISTS idx_pinche_messages_sender_id ON pinche_messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_pinche_messages_receiver_id ON pinche_messages(receiver_id);
CREATE INDEX IF NOT EXISTS idx_pinche_messages_is_read ON pinche_messages(is_read);
CREATE INDEX IF NOT EXISTS idx_pinche_messages_deleted_at ON pinche_messages(deleted_at);

COMMENT ON TABLE pinche_messages IS '拼车行程内消息表';

-- ============================================================
-- 14. pinche_cancels 取消记录表
-- ============================================================
CREATE TABLE IF NOT EXISTS pinche_cancels (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    pinche_id BIGINT NOT NULL,
    booking_id BIGINT,

    cancelled_by BIGINT NOT NULL,
    cancelled_role VARCHAR(16) NOT NULL DEFAULT 'passenger', -- driver/passenger
    cancel_reason VARCHAR(500) NOT NULL DEFAULT '',
    cancel_type VARCHAR(32) NOT NULL DEFAULT 'user',          -- user/system/timeout
    cancel_time TIMESTAMPTZ,

    penalty_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    penalty_paid BOOLEAN NOT NULL DEFAULT FALSE,
    refund_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    refund_status INT NOT NULL DEFAULT 0                      -- 0待退 1已退 2失败
);
CREATE INDEX IF NOT EXISTS idx_pinche_cancels_region_id ON pinche_cancels(region_id);
CREATE INDEX IF NOT EXISTS idx_pinche_cancels_pinche_id ON pinche_cancels(pinche_id);
CREATE INDEX IF NOT EXISTS idx_pinche_cancels_booking_id ON pinche_cancels(booking_id);
CREATE INDEX IF NOT EXISTS idx_pinche_cancels_cancelled_by ON pinche_cancels(cancelled_by);
CREATE INDEX IF NOT EXISTS idx_pinche_cancels_deleted_at ON pinche_cancels(deleted_at);

COMMENT ON TABLE pinche_cancels IS '拼车取消记录表';

-- ============================================================
-- 15. pinche_refunds 退款记录表
-- ============================================================
CREATE TABLE IF NOT EXISTS pinche_refunds (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    payment_id BIGINT NOT NULL,
    booking_id BIGINT NOT NULL,
    pinche_id BIGINT NOT NULL,

    refund_no VARCHAR(32) NOT NULL,
    refund_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    refund_reason VARCHAR(500) NOT NULL DEFAULT '',
    refund_status INT NOT NULL DEFAULT 0, -- 0待退 1已退 2失败
    refund_method VARCHAR(16) NOT NULL DEFAULT 'original', -- original/wechat/alipay/balance

    refunded_at TIMESTAMPTZ,
    third_party_no VARCHAR(64) NOT NULL DEFAULT '',

    operator_id BIGINT,
    handled_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_pinche_refunds_region_id ON pinche_refunds(region_id);
CREATE INDEX IF NOT EXISTS idx_pinche_refunds_payment_id ON pinche_refunds(payment_id);
CREATE INDEX IF NOT EXISTS idx_pinche_refunds_booking_id ON pinche_refunds(booking_id);
CREATE INDEX IF NOT EXISTS idx_pinche_refunds_pinche_id ON pinche_refunds(pinche_id);
CREATE INDEX IF NOT EXISTS idx_pinche_refunds_refund_no ON pinche_refunds(refund_no);
CREATE INDEX IF NOT EXISTS idx_pinche_refunds_status ON pinche_refunds(refund_status);
CREATE INDEX IF NOT EXISTS idx_pinche_refunds_deleted_at ON pinche_refunds(deleted_at);

COMMENT ON TABLE pinche_refunds IS '拼车退款记录表';

-- ============================================================
-- 16. pinche_complaints 投诉表
-- ============================================================
CREATE TABLE IF NOT EXISTS pinche_complaints (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    pinche_id BIGINT NOT NULL,
    booking_id BIGINT,
    trip_id BIGINT,

    complainant_id BIGINT NOT NULL,
    complainant_name VARCHAR(50) NOT NULL DEFAULT '',
    respondent_id BIGINT NOT NULL,
    respondent_name VARCHAR(50) NOT NULL DEFAULT '',

    complaint_type VARCHAR(32) NOT NULL DEFAULT '',
    complaint_reason VARCHAR(500) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    evidence_images JSONB,

    status INT NOT NULL DEFAULT 0, -- 0待处理 1处理中 2已处理 3已驳回
    handler_id BIGINT,
    handler_name VARCHAR(50) NOT NULL DEFAULT '',
    handle_result VARCHAR(500) NOT NULL DEFAULT '',
    handled_at TIMESTAMPTZ,

    penalty_type VARCHAR(32) NOT NULL DEFAULT '',
    penalty_user_id BIGINT,
    sla_deadline TIMESTAMPTZ,

    appealed_at TIMESTAMPTZ,
    appeal_result VARCHAR(500) NOT NULL DEFAULT '',
    appeal_handler_id BIGINT
);
CREATE INDEX IF NOT EXISTS idx_pinche_complaints_region_id ON pinche_complaints(region_id);
CREATE INDEX IF NOT EXISTS idx_pinche_complaints_pinche_id ON pinche_complaints(pinche_id);
CREATE INDEX IF NOT EXISTS idx_pinche_complaints_booking_id ON pinche_complaints(booking_id);
CREATE INDEX IF NOT EXISTS idx_pinche_complaints_complainant_id ON pinche_complaints(complainant_id);
CREATE INDEX IF NOT EXISTS idx_pinche_complaints_respondent_id ON pinche_complaints(respondent_id);
CREATE INDEX IF NOT EXISTS idx_pinche_complaints_status ON pinche_complaints(status);
CREATE INDEX IF NOT EXISTS idx_pinche_complaints_deleted_at ON pinche_complaints(deleted_at);

COMMENT ON TABLE pinche_complaints IS '拼车投诉表';

-- ============================================================
-- 17. pinche_audit_rules 审核规则表
-- ============================================================
CREATE TABLE IF NOT EXISTS pinche_audit_rules (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    rule_name VARCHAR(128) NOT NULL DEFAULT '',
    rule_type VARCHAR(32) NOT NULL DEFAULT '',
    rule_code VARCHAR(64) NOT NULL DEFAULT '',
    description VARCHAR(500) NOT NULL DEFAULT '',

    threshold JSONB,
    action VARCHAR(32) NOT NULL DEFAULT 'manual_review', -- pass/reject/manual_review
    priority INT NOT NULL DEFAULT 0,

    status INT NOT NULL DEFAULT 1, -- 0禁用 1启用
    hit_count INT NOT NULL DEFAULT 0,
    last_hit_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_pinche_audit_rules_region_id ON pinche_audit_rules(region_id);
CREATE INDEX IF NOT EXISTS idx_pinche_audit_rules_rule_type ON pinche_audit_rules(rule_type);
CREATE INDEX IF NOT EXISTS idx_pinche_audit_rules_rule_code ON pinche_audit_rules(rule_code);
CREATE INDEX IF NOT EXISTS idx_pinche_audit_rules_status ON pinche_audit_rules(status);
CREATE INDEX IF NOT EXISTS idx_pinche_audit_rules_priority ON pinche_audit_rules(priority);
CREATE INDEX IF NOT EXISTS idx_pinche_audit_rules_deleted_at ON pinche_audit_rules(deleted_at);

COMMENT ON TABLE pinche_audit_rules IS '拼车审核规则表';

-- ============================================================
-- 18. pinche_statistics 统计表
-- ============================================================
CREATE TABLE IF NOT EXISTS pinche_statistics (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    region_id BIGINT NOT NULL DEFAULT 1,

    stat_date DATE NOT NULL,
    stat_type VARCHAR(16) NOT NULL DEFAULT 'daily', -- daily/weekly/monthly/total
    user_id BIGINT,

    total_trips INT NOT NULL DEFAULT 0,
    completed_trips INT NOT NULL DEFAULT 0,
    cancelled_trips INT NOT NULL DEFAULT 0,
    total_bookings INT NOT NULL DEFAULT 0,
    completed_bookings INT NOT NULL DEFAULT 0,

    total_revenue DECIMAL(12,2) NOT NULL DEFAULT 0,
    total_refund DECIMAL(12,2) NOT NULL DEFAULT 0,
    avg_rating DECIMAL(3,2) NOT NULL DEFAULT 0,
    avg_price DECIMAL(12,2) NOT NULL DEFAULT 0,

    total_distance DECIMAL(10,2) NOT NULL DEFAULT 0,
    total_duration INT NOT NULL DEFAULT 0,

    total_passengers INT NOT NULL DEFAULT 0,
    total_drivers INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_pinche_statistics_region_id ON pinche_statistics(region_id);
CREATE INDEX IF NOT EXISTS idx_pinche_statistics_stat_date ON pinche_statistics(stat_date);
CREATE INDEX IF NOT EXISTS idx_pinche_statistics_stat_type ON pinche_statistics(stat_type);
CREATE INDEX IF NOT EXISTS idx_pinche_statistics_user_id ON pinche_statistics(user_id);
CREATE INDEX IF NOT EXISTS idx_pinche_statistics_deleted_at ON pinche_statistics(deleted_at);

COMMENT ON TABLE pinche_statistics IS '拼车统计表';

-- ============================================================
-- updated_at 触发器（依赖 001_p0_baseline.sql 中的 update_updated_at_column 函数）
-- ============================================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'update_updated_at_column') THEN
        DROP TRIGGER IF EXISTS trg_pinches_updated_at ON pinches; CREATE TRIGGER trg_pinches_updated_at BEFORE UPDATE ON pinches FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_pinche_routes_updated_at ON pinche_routes; CREATE TRIGGER trg_pinche_routes_updated_at BEFORE UPDATE ON pinche_routes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_pinche_bookings_updated_at ON pinche_bookings; CREATE TRIGGER trg_pinche_bookings_updated_at BEFORE UPDATE ON pinche_bookings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_pinche_drivers_updated_at ON pinche_drivers; CREATE TRIGGER trg_pinche_drivers_updated_at BEFORE UPDATE ON pinche_drivers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_pinche_vehicles_updated_at ON pinche_vehicles; CREATE TRIGGER trg_pinche_vehicles_updated_at BEFORE UPDATE ON pinche_vehicles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_pinche_insurances_updated_at ON pinche_insurances; CREATE TRIGGER trg_pinche_insurances_updated_at BEFORE UPDATE ON pinche_insurances FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_pinche_payments_updated_at ON pinche_payments; CREATE TRIGGER trg_pinche_payments_updated_at BEFORE UPDATE ON pinche_payments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_pinche_ratings_updated_at ON pinche_ratings; CREATE TRIGGER trg_pinche_ratings_updated_at BEFORE UPDATE ON pinche_ratings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_pinche_emergencies_updated_at ON pinche_emergencies; CREATE TRIGGER trg_pinche_emergencies_updated_at BEFORE UPDATE ON pinche_emergencies FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_pinche_trips_updated_at ON pinche_trips; CREATE TRIGGER trg_pinche_trips_updated_at BEFORE UPDATE ON pinche_trips FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_pinche_route_favorites_updated_at ON pinche_route_favorites; CREATE TRIGGER trg_pinche_route_favorites_updated_at BEFORE UPDATE ON pinche_route_favorites FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_pinche_messages_updated_at ON pinche_messages; CREATE TRIGGER trg_pinche_messages_updated_at BEFORE UPDATE ON pinche_messages FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_pinche_cancels_updated_at ON pinche_cancels; CREATE TRIGGER trg_pinche_cancels_updated_at BEFORE UPDATE ON pinche_cancels FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_pinche_refunds_updated_at ON pinche_refunds; CREATE TRIGGER trg_pinche_refunds_updated_at BEFORE UPDATE ON pinche_refunds FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_pinche_complaints_updated_at ON pinche_complaints; CREATE TRIGGER trg_pinche_complaints_updated_at BEFORE UPDATE ON pinche_complaints FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_pinche_audit_rules_updated_at ON pinche_audit_rules; CREATE TRIGGER trg_pinche_audit_rules_updated_at BEFORE UPDATE ON pinche_audit_rules FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
        DROP TRIGGER IF EXISTS trg_pinche_statistics_updated_at ON pinche_statistics; CREATE TRIGGER trg_pinche_statistics_updated_at BEFORE UPDATE ON pinche_statistics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;
