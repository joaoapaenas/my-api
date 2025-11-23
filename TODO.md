
---

### 1. Fluxo Operacional (Core Logic)

#### A. Carregar o Ciclo Ativo (Round Robin)
O app precisa saber a ordem das matérias. Esta query busca os itens do ciclo ativo ordenados.
*   **Uso:** Ao abrir o app ou ao clicar em "Estudar". O Backend (Go) deve manter um ponteiro (state) de qual `order_index` é o atual.

```sql
SELECT 
    ci.id AS cycle_item_id,
    ci.order_index,
    ci.planned_duration_minutes,
    s.id AS subject_id,
    s.name AS subject_name,
    s.color_hex
FROM cycle_items ci
JOIN study_cycles sc ON ci.cycle_id = sc.id
JOIN subjects s ON ci.subject_id = s.id
WHERE sc.is_active = 1 
  AND sc.deleted_at IS NULL
ORDER BY ci.order_index ASC;
```

#### B. Recuperar Sessão em Aberto (Crash Recovery)
Se o app fechar inesperadamente ou o celular desligar, precisamos recuperar a sessão que não tem `finished_at`.
*   **Uso:** Executar no `init()` do app.
*   **Mitigação:** Se houver uma sessão aberta há mais de 24h, o backend deve sugerir descartá-la ou fechá-la automaticamente.

```sql
SELECT 
    ss.id,
    ss.started_at,
    s.name AS subject_name
FROM study_sessions ss
JOIN subjects s ON ss.subject_id = s.id
WHERE ss.finished_at IS NULL
ORDER BY ss.started_at DESC
LIMIT 1;
```

---

### 2. Analytics de Tempo (Time Tracking)

#### C. Relatório de Tempo Líquido por Matéria (Semanal/Geral)
Quanto tempo o usuário realmente estudou de cada matéria?
*   **Engineering Note:** `COALESCE` garante que não retornemos NULL.

```sql
SELECT 
    s.name AS subject_name,
    COUNT(ss.id) AS sessions_count,
    -- Soma do tempo líquido em horas (com 2 casas decimais)
    ROUND(SUM(ss.net_duration_seconds) / 3600.0, 2) AS total_hours_net
FROM study_sessions ss
JOIN subjects s ON ss.subject_id = s.id
WHERE ss.finished_at IS NOT NULL
-- Filtro de Data (Ex: Últimos 7 dias) - O Go deve injetar a string da data
-- AND ss.started_at >= '2023-10-01T00:00:00Z' 
GROUP BY s.id, s.name
ORDER BY total_hours_net DESC;
```

---

### 3. Analytics de Performance (Questões)

#### D. Aproveitamento Global por Matéria
Calcula a porcentagem de acertos.
*   **Defensive SQL:** Utilizo `NULLIF` para evitar divisão por zero e multiplico por `1.0` para forçar a divisão de ponto flutuante (SQLite trata divisão de inteiros como inteiro).

```sql
SELECT 
    s.name AS subject_name,
    SUM(el.questions_count) AS total_questions,
    SUM(el.correct_count) AS total_correct,
    -- Cálculo de Porcentagem
    ROUND(
        (SUM(el.correct_count) * 100.0) / NULLIF(SUM(el.questions_count), 0), 
        2
    ) AS accuracy_percentage
FROM exercise_logs el
JOIN subjects s ON el.subject_id = s.id
GROUP BY s.id, s.name
HAVING total_questions > 0 -- Retornar apenas matérias com questões feitas
ORDER BY accuracy_percentage ASC; -- Mostra o ponto fraco primeiro
```

#### E. Análise de Pontos Fracos (Drill-down por Tópico)
O usuário vai mal em "Direito Constitucional". Onde exatamente?
*   **Uso:** Clicar em uma matéria para ver o detalhe dos tópicos.

```sql
SELECT 
    t.name AS topic_name,
    SUM(el.questions_count) AS total_questions,
    ROUND(
        (SUM(el.correct_count) * 100.0) / NULLIF(SUM(el.questions_count), 0), 
        2
    ) AS accuracy_percentage
FROM exercise_logs el
JOIN topics t ON el.topic_id = t.id
WHERE el.subject_id = ? -- Injetar ID da matéria (ex: UUID do Direito Const.)
GROUP BY t.id, t.name
ORDER BY accuracy_percentage ASC;
```

---

### 4. Consistência (Habit Building)

#### F. Mapa de Calor (Atividade Diária)
Para preencher aquele calendário estilo GitHub (Heatmap).
*   **SQL Trick:** `strftime('%Y-%m-%d', ...)` trunca o timestamp para o dia.

```sql
SELECT 
    strftime('%Y-%m-%d', started_at) AS study_date,
    COUNT(DISTINCT id) AS sessions_count,
    SUM(net_duration_seconds) AS total_seconds
FROM study_sessions
WHERE finished_at IS NOT NULL
GROUP BY study_date
ORDER BY study_date DESC
LIMIT 30; -- Últimos 30 dias com atividade
```