import http from 'k6/http';
import {check, sleep} from 'k6';

export const options = {
    // 关键：通过足够的VU产生高TPS（需根据系统响应时间调整）
    vus: 1000, // 初始虚拟用户数（可逐步增加至2000）
    duration: '60s', // 测试持续时间（足够长以达到稳定状态）

    // 流量控制：确保每秒请求数稳定在5000左右
    rps: 5000, // 限制每秒请求数（k6 v0.38+支持，确保版本兼容）

    // 性能阈值（根据系统预期设置）
    thresholds: {
        http_req_duration: ['p(95)<50'], // 95%请求延迟<50ms
        http_req_failed: ['rate<0.01'], // 失败率<1%
        http_reqs: ['rate>5000'], // 确保TPS>5000
    },
};

// 基础配置（复用之前的正确参数）
const POST_URL = 'http://localhost:8080/articles/list';
// JwtToken要从前端拿
const JWT_TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTQxOTg2NTMsIlVpZCI6MSwiU3NpZCI6ImMwZmIzNjc1LWI0NDctNGI5OC1hNDRjLTdkNzIzZjMyZDQ0OSIsIlVzZXJBZ2VudCI6Ik1vemlsbGEvNS4wIChXaW5kb3dzIE5UIDEwLjA7IFdpbjY0OyB4NjQpIEFwcGxlV2ViS2l0LzUzNy4zNiAoS0hUTUwsIGxpa2UgR2Vja28pIENocm9tZS8xMzguMC4wLjAgU2FmYXJpLzUzNy4zNiJ9.B0PBSk3SUCuGt5p_hM6HrK1fyrtWz_pwHRU3YBFofv8';
const POST_BODY = JSON.stringify({ Offset: 0, Limit: 10 });

export default function () {
    const headers = {
        'Authorization': `Bearer ${JWT_TOKEN}`,
        'Content-Type': 'application/json',
        'Origin': 'http://localhost:3000',
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36'
    };

    // 发送POST请求
    const res = http.post(POST_URL, POST_BODY, { headers });

    // 验证请求成功
    check(res, {
        "认证成功2xx": (r) => r.status >= 200 && r.status < 300,
    });

    // 可选：微调请求频率（若rps控制不生效，可加极短延迟）
    // sleep(0.001); // 1ms延迟，避免请求过于密集导致系统瞬间过载

}