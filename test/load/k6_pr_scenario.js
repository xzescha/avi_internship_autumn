import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

export const biz_fail_rate = new Rate('biz_fail_rate');

export const options = {
    scenarios: {
        pr_flow: {
            executor: 'constant-arrival-rate',
            rate: 5,              // 5 запросов в секунду
            timeUnit: '1s',
            duration: '60s',
            preAllocatedVUs: 10,
            maxVUs: 20,
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<300'], // SLI по времени
        biz_fail_rate: ['rate<0.001'],    // SLI успешности 99.9%
    },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

function randomInt(min, max) {
    return Math.floor(Math.random() * (max - min + 1)) + min;
}

export function setup() {
    const teamCount = 20;
    const usersPerTeam = 10;
    let userId = 1;

    for (let t = 1; t <= teamCount; t++) {
        const teamName = `team_${t}`;
        const members = [];

        for (let i = 0; i < usersPerTeam; i++) {
            members.push({
                user_id: `u${userId}`,
                username: `user_${userId}`,
                is_active: true,
            });
            userId++;
        }

        const res = http.post(
            `${BASE_URL}/team/add`,
            JSON.stringify({
                team_name: teamName,
                members: members,
            }),
            {
                headers: { 'Content-Type': 'application/json' },
                tags: { name: 'team_add' },
            }
        );

        // ожидаем либо успех, либо "уже существует"
        const ok = res.status === 201 || res.status === 400;
        biz_fail_rate.add(!ok || res.status >= 500);

        check(res, {
            'team created or already exists': () => ok,
        });
    }

    return { totalUsers: 200 };
}

export default function (data) {
    const totalUsers = data.totalUsers || 200;

    const authorID = `u${randomInt(1, totalUsers)}`;
    const prID = `pr-${__VU}-${__ITER}`;

    // 1) create PR
    let res = http.post(
        `${BASE_URL}/pullRequest/create`,
        JSON.stringify({
            pull_request_id: prID,
            pull_request_name: `Feature ${__VU}-${__ITER}`,
            author_id: authorID,
        }),
        {
            headers: { 'Content-Type': 'application/json' },
            tags: { name: 'pr_create' },
        }
    );

    // ожидаем 201, 409 или 404 (PR уже существует / конфликт/ не найден пользователь или Pull Request)
    let ok = res.status === 201 || res.status === 409 || res.status === 404; // Это все подходящие под логику проекта статусы ответа с сервера
    biz_fail_rate.add(!ok || res.status >= 500);

    check(res, {
        'create PR 201 or conflict': () => ok,
    });

    // 2) reassign reviewer (если есть хотя бы один ревьювер)
    const body = JSON.parse(res.body || '{}');
    if (body.pr && body.pr.assigned_reviewers && body.pr.assigned_reviewers.length > 0) {
        const oldReviewerID = body.pr.assigned_reviewers[0];

        const resReassign = http.post(
            `${BASE_URL}/pullRequest/reassign`,
            JSON.stringify({
                pull_request_id: prID,
                old_user_id: oldReviewerID,
            }),
            {
                headers: { 'Content-Type': 'application/json' },
                tags: { name: 'pr_reassign' },
            }
        );

        // ожидаем 200 или 409 (not_assigned/no_candidate и т.п.)
        ok = resReassign.status === 200 || resReassign.status === 409;
        biz_fail_rate.add(!ok || resReassign.status >= 500);

        check(resReassign, {
            'reassign ok or business-conflict': () => ok,
        });
    }

    // 3) периодически дергаем статистику
    if (__ITER % 5 === 0) {
        const statsRes = http.get(`${BASE_URL}/stats/assignments`, {
            tags: { name: 'stats_assignments' },
        });

        ok = statsRes.status === 200;
        biz_fail_rate.add(!ok || statsRes.status >= 500);

        check(statsRes, { 'stats 200': () => ok });
    }

    sleep(0.1);
}
