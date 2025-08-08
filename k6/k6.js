import http from 'k6/http';
import {check, sleep} from 'k6'; // 导入检查和延迟函数

// 压测配置：定义虚拟用户数和测试时长
export let options = {
    vus: 10,  // 并发虚拟用户数（根据需求调整，如10、50、100等）
    duration: '5s',  // 测试持续时间（如30秒、5分钟等）

    // 可选：分阶段压测（更贴近真实场景）
    // stages: [
    //   { duration: '10s', target: 10 },  // 10秒内逐步增加到10个VUs
    //   { duration: '20s', target: 10 },  // 保持10个VUs持续20秒
    //   { duration: '10s', target: 0 },   // 10秒内逐步减少到0个VUs
    // ],
};

const url = "http://localhost:8080/hello";

export default function () {
    const data = { name: "Tom" };

    // 发送POST请求
    const res = http.get(
        url,
        JSON.stringify(data),
        { headers: { 'Content-Type': 'application/json' } }
    );

    // 检查响应是否符合预期（替代responseCallback的方式，更灵活）
    check(res, {
        "状态码为200": (r) => r.status === 200,  // 验证HTTP状态码
        "响应时间<1s": (r) => r.timings.duration < 1000,  // 验证响应速度
        // 可添加更多检查，如响应内容包含特定字段等
        // "响应包含name字段": (r) => JSON.parse(r.body).name === "Tom",
    });

    // 模拟用户思考时间（可选，使压测更接近真实用户行为）
    sleep(1);  // 每个请求后等待1秒（根据实际场景调整）
}
