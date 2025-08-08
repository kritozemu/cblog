import http from 'k6/http';
import {check} from 'k6';

export const options = {
    vus: 5,
    duration: '10s',
};

// 1. 确认与Postman完全一致的URL和方法（POST）
const POST_URL = 'http://localhost:8080/articles/list';
// 2. 从Postman的Authorization头中复制的完整token（确保无修改）
// JwtToken从前端拿
const JWT_TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTQxOTg2NTMsIlVpZCI6MSwiU3NpZCI6ImMwZmIzNjc1LWI0NDctNGI5OC1hNDRjLTdkNzIzZjMyZDQ0OSIsIlVzZXJBZ2VudCI6Ik1vemlsbGEvNS4wIChXaW5kb3dzIE5UIDEwLjA7IFdpbjY0OyB4NjQpIEFwcGxlV2ViS2l0LzUzNy4zNiAoS0hUTUwsIGxpa2UgR2Vja28pIENocm9tZS8xMzguMC4wLjAgU2FmYXJpLzUzNy4zNiJ9.B0PBSk3SUCuGt5p_hM6HrK1fyrtWz_pwHRU3YBFofv8';
// 3. 与Postman完全一致的请求体（Body）
const POST_BODY = JSON.stringify({
    Offset: 0,
    Limit: 10
    // 补充Postman中Body的其他参数（如有）
});

export default function () {
    // 完全复刻Postman的请求头（包括认证和跨域信息）
    const headers = {
        // 核心认证头（严格格式：Bearer + 空格 + token）
        'Authorization': `Bearer ${JWT_TOKEN}`,
        // POST请求必需的Content-Type（与Postman一致）
        'Content-Type': 'application/json',
        // 跨域相关头（后端可能结合这些验证来源合法性）
        'Origin': 'http://localhost:3000',
        'Referer': 'http://localhost:3000/',
        // 浏览器标识头（确保后端识别为合法客户端）
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36',
        'sec-ch-ua': '"Not)A;Brand";v="8", "Chromium";v="138", "Google Chrome";v="138"',
        'sec-ch-ua-platform': '"Windows"',
        'sec-fetch-mode': 'cors'
    };

    // 发送POST请求（与Postman方法一致）
    const res = http.post(POST_URL, POST_BODY, { headers });

    // // 详细日志（辅助定位认证失败原因）
    // console.log(`状态码: ${res.status}`);
    // console.log(`认证头: ${headers.Authorization.substring(0, 50)}...`); // 验证认证头格式
    // console.log(`响应头: ${JSON.stringify(res.headers)}`); // 查看后端是否返回认证相关提示

    check(res, {
        '认证成功(2xx)': (r) => r.status >= 200 && r.status < 300,
    });
}