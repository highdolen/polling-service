const API_URL = '';

const poll = document.getElementById('poll');
const question = document.getElementById('question');
const options = document.getElementById('options');
const form = document.getElementById('vote-form');
const button = document.getElementById('vote-button');
const message = document.getElementById('message');

let pollData = null;

async function loadPoll() {
    console.log('loadPoll started');

    try {
        const response = await fetch(
            `${API_URL}/api/v1/polls/active`,
        );

        console.log(
            'GET active poll:',
            response.status,
            response.statusText,
        );

        if (response.status === 404) {
            showMessage('Сейчас активных опросов нет.');
            return;
        }

        if (!response.ok) {
            throw new Error(
                `Не удалось загрузить опрос. HTTP ${response.status}`,
            );
        }

        pollData = await response.json();

        console.log('Active poll response:', pollData);

        renderPoll();
    } catch (error) {
        console.error('Load poll error:', error);

        showMessage(
            'Не удалось загрузить опрос. Попробуйте обновить страницу.',
        );
    }
}

function renderPoll() {
    const pollQuestion =
        pollData.Question ?? pollData.question;

    const pollOptions =
        pollData.Options ?? pollData.options;

    const pollType =
        pollData.Type ?? pollData.type;

    if (!pollQuestion) {
        throw new Error(
            'У опроса отсутствует вопрос.',
        );
    }

    if (!pollOptions || pollOptions.length === 0) {
        throw new Error(
            'У опроса отсутствуют варианты ответа.',
        );
    }

    question.textContent = pollQuestion;

    options.innerHTML = '';

    const inputType =
        pollType === 'multiple'
            ? 'checkbox'
            : 'radio';

    pollOptions.forEach((option) => {
        const optionID =
            option.ID ?? option.id;

        const optionText =
            option.Text ?? option.text;

        const label = document.createElement('label');

        label.className = 'option';

        const input = document.createElement('input');

        input.type = inputType;
        input.name = 'option';
        input.value = optionID;

        const text = document.createElement('span');

        text.textContent = optionText;

        label.appendChild(input);
        label.appendChild(text);

        options.appendChild(label);
    });

    poll.classList.remove('hidden');
}

form.addEventListener('submit', async (event) => {
    event.preventDefault();

    if (!pollData) {
        return;
    }

    const selectedOptions = [
        ...document.querySelectorAll(
            'input[name="option"]:checked',
        ),
    ];

    if (selectedOptions.length === 0) {
        showMessage('Выберите вариант ответа.');
        return;
    }

    const optionIDs = selectedOptions.map(
        (input) => Number(input.value),
    );

    const pollID =
        pollData.ID ?? pollData.id;

    if (!pollID) {
        showMessage(
            'Не удалось определить ID опроса.',
        );
        return;
    }

    button.disabled = true;

    try {
        const response = await fetch(
            `${API_URL}/api/v1/polls/${pollID}/vote`,
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    option_ids: optionIDs,
                }),
            },
        );

        console.log(
            'POST vote:',
            response.status,
            response.statusText,
        );

        if (response.status === 201) {
            showMessage('Ваш голос принят!');
            form.reset();
            return;
        }

        if (response.status === 409) {
            showMessage(
                'Вы уже голосовали в этом опросе.',
            );
            return;
        }

        if (response.status === 400) {
            showMessage(
                'Голосование сейчас недоступно.',
            );
            return;
        }

        if (response.status === 404) {
            showMessage(
                'Опрос больше недоступен.',
            );
            return;
        }

        showMessage(
            `Произошла ошибка. HTTP ${response.status}`,
        );
    } catch (error) {
        console.error('Vote error:', error);

        showMessage(
            'Не удалось отправить голос.',
        );
    } finally {
        button.disabled = false;
    }
});

function showMessage(text) {
    message.textContent = text;
    message.classList.remove('hidden');
}

loadPoll();
 


