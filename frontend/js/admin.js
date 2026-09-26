const API_URL = '';

const createForm = document.getElementById('create-poll-form');
const addOptionButton = document.getElementById('add-option');
const optionsContainer = document.getElementById('options');
const createMessage = document.getElementById('create-message');
const loadPollsButton = document.getElementById('load-polls');
const pollsList = document.getElementById('polls-list');
const resultsContainer = document.getElementById('results');

let adminToken = '';

createForm.addEventListener('submit', async (event) => {
    event.preventDefault();

    adminToken = document.getElementById('admin-token').value;

    const optionInputs = [
        ...document.querySelectorAll('.option-input'),
    ];

    const optionValues = optionInputs
        .map((input) => input.value.trim())
        .filter((value) => value !== '');

    if (optionValues.length < 2) {
        showCreateMessage(
            'Добавьте минимум два варианта ответа.',
        );
        return;
    }

    const startsAt = moscowTimeToUTC(
        document.getElementById('starts-at').value,
    );

    const endsAt = moscowTimeToUTC(
        document.getElementById('ends-at').value,
    );

    if (!startsAt || !endsAt) {
        showCreateMessage(
            'Укажите корректное время начала и окончания.',
        );
        return;
    }

    if (new Date(startsAt) >= new Date(endsAt)) {
        showCreateMessage(
            'Время окончания должно быть позже времени начала.',
        );
        return;
    }

    const payload = {
        question: document.getElementById('poll-question').value,
        type: document.getElementById('poll-type').value,
        starts_at: startsAt,
        ends_at: endsAt,
        options: optionValues,
    };

    try {
        const response = await fetch(
            `${API_URL}/api/v1/admin/polls`,
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-Admin-Token': adminToken,
                },
                body: JSON.stringify(payload),
            },
        );

        if (response.status === 401) {
            showCreateMessage('Неверный admin token.');
            return;
        }

        if (response.status === 409) {
            showCreateMessage('В указанное время уже существует другой опрос.');
            return;
        }

        if (!response.ok) {
            showCreateMessage(
                'Не удалось создать опрос.',
            );
            return;
        }

        const poll = await response.json();

        showCreateMessage(
            `Опрос создан. ID: ${poll.id}`,
        );

        createForm.reset();

        optionsContainer.innerHTML = `
            <input
                type="text"
                class="option-input"
                placeholder="Вариант 1"
                required
            >
            <input
                type="text"
                class="option-input"
                placeholder="Вариант 2"
                required
            >
        `;
    } catch (error) {
        showCreateMessage(
            'Не удалось подключиться к серверу.',
        );
    }
});

addOptionButton.addEventListener('click', () => {
    const count =
        document.querySelectorAll('.option-input').length + 1;

    const input = document.createElement('input');

    input.type = 'text';
    input.className = 'option-input';
    input.placeholder = `Вариант ${count}`;
    input.required = true;

    optionsContainer.appendChild(input);
});

loadPollsButton.addEventListener('click', async () => {
    if (!adminToken) {
        adminToken = document.getElementById('admin-token').value;
    }

    if (!adminToken) {
        pollsList.textContent = 'Введите admin token.';
        return;
    }

    try {
        const response = await fetch(
            `${API_URL}/api/v1/admin/polls`,
            {
                headers: {
                    'X-Admin-Token': adminToken,
                },
            },
        );

        if (response.status === 401) {
            pollsList.textContent =
                'Неверный admin token.';
            return;
        }

        if (!response.ok) {
            pollsList.textContent =
                'Не удалось загрузить опросы.';
            return;
        }

        const polls = await response.json();

        renderPolls(polls);
    } catch (error) {
        pollsList.textContent =
            'Не удалось подключиться к серверу.';
    }
});

function renderPolls(polls) {
    pollsList.innerHTML = '';

    if (polls.length === 0) {
        pollsList.textContent = 'Опросов пока нет.';
        return;
    }

    polls.forEach((poll) => {
        const item = document.createElement('div');

        item.className = 'poll-item';

        const title = document.createElement('h3');

        title.textContent =
            `#${poll.ID} — ${poll.Question}`;

        const info = document.createElement('p');

        info.textContent =
            `Тип: ${poll.Type}, статус: ${poll.Status}`;

        const time = document.createElement('p');

        time.textContent =
            `Начало: ${formatMoscowTime(poll.StartsAt)} | ` +
            `Окончание: ${formatMoscowTime(poll.EndsAt)}`;

        const button = document.createElement('button');

        button.textContent =
            'Посмотреть результаты';

        button.addEventListener('click', () => {
            loadResults(poll.ID);
        });

        item.appendChild(title);
        item.appendChild(info);
        item.appendChild(time);
        item.appendChild(button);

        pollsList.appendChild(item);
    });
}

async function loadResults(pollID) {
    try {
        const response = await fetch(
            `${API_URL}/api/v1/admin/polls/${pollID}/results`,
            {
                headers: {
                    'X-Admin-Token': adminToken,
                },
            },
        );

        if (!response.ok) {
            resultsContainer.textContent =
                'Не удалось загрузить результаты.';
            return;
        }

        const results = await response.json();

        renderResults(results);
    } catch (error) {
        resultsContainer.textContent =
            'Не удалось подключиться к серверу.';
    }
}

function renderResults(results) {
    resultsContainer.innerHTML = '';

    if (results.length === 0) {
        resultsContainer.textContent =
            'Результатов пока нет.';
        return;
    }

    const maxVotes = Math.max(
        ...results.map((result) => result.VotesCount),
        1,
    );

    results.forEach((result, index) => {
        const item = document.createElement('div');

        item.className = 'result-item';

        const text = document.createElement('div');

        text.textContent =
            `Вариант ${index + 1}: ` +
            `${result.VotesCount} голосов`;

        const bar = document.createElement('div');

        bar.className = 'result-bar';

        const value = document.createElement('div');

        value.className = 'result-value';

        value.style.width =
            `${(result.VotesCount / maxVotes) * 100}%`;

        bar.appendChild(value);
        item.appendChild(text);
        item.appendChild(bar);

        resultsContainer.appendChild(item);
    });
}

function showCreateMessage(text) {
    createMessage.textContent = text;
    createMessage.classList.remove('hidden');
}

function moscowTimeToUTC(value) {
    if (!value) {
        return null;
    }

    const [datePart, timePart] = value.split('T');

    if (!datePart || !timePart) {
        return null;
    }

    const [year, month, day] = datePart
        .split('-')
        .map(Number);

    const [hours, minutes] = timePart
        .split(':')
        .map(Number);

    // Москва = UTC+3.
    const utcDate = new Date(
        Date.UTC(
            year,
            month - 1,
            day,
            hours - 3,
            minutes,
        ),
    );

    return utcDate.toISOString();
}

function formatMoscowTime(value) {
    const date = new Date(value);

    return date.toLocaleString('ru-RU', {
        timeZone: 'Europe/Moscow',
        dateStyle: 'short',
        timeStyle: 'short',
    });
}