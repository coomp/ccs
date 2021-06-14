-- checks if water_no already exists
local order_no_key = KEYS[1]
local function is_order_existed(order_no_key)
    -- body
    local reply = redis.pcall('GET', order_no_key)
    if reply ~= false then return  110001 end
end

-- 自增
return redis.call("ZADD",  "ccs@")
