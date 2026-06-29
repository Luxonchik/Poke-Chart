# CLAUDE.md — PokeTrack Frontend

Этот файл задаёт правила архитектуры и стиля для проекта **PokeTrack** —
фронтенда для отслеживания цен на покемон-карточки с корейских барахолок.

Claude должен **строго следовать этим правилам** при генерации, рефакторинге
и ревью кода в этом репозитории. Если задача противоречит правилам ниже —
сначала указать на конфликт, а не молча отступать от структуры.

## Стек

- **Next.js** (App Router)
- **TypeScript** (строгий режим, без `any` без крайней необходимости)
- **TanStack Query** — серверный стейт / кэширование запросов
- **Zustand** — глобальный UI-стейт (фильтры, модалки, тоглы)
- **Tailwind CSS**
- **shadcn/ui**
- **next-intl** — мультиязычность (KO/EN)
- Backend: отдельный REST API (Go/Node), фронтенд только потребляет его

## Структура проекта

Актуальная структура (`web/`):

```
app/
  [locale]/                  # ko / en — next-intl locale-сегмент, ОБЯЗАТЕЛЕН для всех роутов
    layout.tsx
    page.tsx
    (другие роуты по той же схеме: app/[locale]/<route>/page.tsx)
  favicon.ico
  globals.css

components/
  ui/                        # shadcn-компоненты как есть, не трогать руками без причины
  cards/                     # компоненты домена "карточки"
    InfiniteCardGrid.tsx
    PokemonCard.tsx
  layout/                    # компоненты каркаса приложения
    Header.tsx
  providers/                 # провайдеры верхнего уровня
    QueryProvider.tsx
  features/                  # самостоятельные секции/фичи страниц (не переиспользуемые виджеты)
    landing/
      Trending.tsx

lib/
  api/                       # fetch-функции к REST backend, по файлу на домен
    cards.ts
    marketplaces.ts
  queries/                   # хуки TanStack Query (useQuery/useMutation), по файлу на домен
    useCards.ts
    useMarketplaces.ts
  store.ts                   # все Zustand-сторы (один файл, несколько хуков-сторов внутри)
  types.ts                   # общие типы/интерфейсы (Card, Marketplace, PriceEntry...)
  constants.ts
  utils.ts

i18n/                        # конфигурация next-intl
  routing.ts
  request.ts
messages/                    # словари переводов
  ko.json
  en.json
proxy.ts                     # next-intl middleware (определение/редирект локали)
```

> Сейчас `lib/api.ts` и `lib/queries/` — один файл без разбивки по доменам.
> Целевая схема (выше) — разбить `api.ts` на `lib/api/<domain>.ts` и
> завести `lib/queries/<domain>.ts` с хуками. Новый код пиши уже по целевой
> схеме; старый переносить можно постепенно, не блокируя текущие задачи.

### Правило размещения компонентов

- `components/ui/*` — только сгенерированные/импортированные shadcn-компоненты.
  Кастомная бизнес-логика туда не добавляется.
- `components/<domain>/*` — компоненты, привязанные к домену данных
  (`cards/`, `marketplace/` и т.п.), используются на нескольких страницах.
- `components/layout/*` — каркас приложения: Header, Footer, Sidebar, Nav.
- `components/providers/*` — провайдеры контекста (QueryProvider и
  аналогичные), монтируются в `app/[locale]/layout.tsx`.
- `components/features/<feature-name>/*` — крупные самодостаточные секции
  конкретной страницы, которые не переиспользуются в другом месте
  (например `features/landing/Trending.tsx`). Если фича начинает
  переиспользоваться на 2+ страницах — переносить в `components/<domain>/`.

## Слой данных: TanStack Query

- **Все** обращения к backend идут через `lib/api/<domain>.ts` — это чистые
  async функции (`fetch`), без React-хуков внутри. Один файл = один домен
  (`cards.ts`, `marketplaces.ts`, `prices.ts`).
- **Все** `useQuery`/`useMutation` живут в `lib/queries/<domain>.ts`, по
  одному файлу на домен, имя файла = `use<Domain>.ts`. Компоненты импортируют
  готовый хук, а не вызывают `useQuery` напрямую.
- Ключи запросов (`queryKey`) централизованы — фабрика ключей в том же файле
  или в `lib/queries/keys.ts`, без магических строк-литералов в компонентах.

```ts
// lib/api/cards.ts
export async function fetchCards(params: CardsParams): Promise<Card[]> {
  const res = await fetch(`${API_URL}/cards?${toQueryString(params)}`);
  if (!res.ok) throw new Error("Failed to fetch cards");
  return res.json();
}

// lib/queries/useCards.ts
export function useCards(params: CardsParams) {
  return useQuery({
    queryKey: ["cards", params],
    queryFn: () => fetchCards(params),
  });
}

// components/cards/InfiniteCardGrid.tsx
const { data, isLoading } = useCards(filters);
```

- Мутации с инвалидацией кэша: `onSuccess` вызывает
  `queryClient.invalidateQueries({ queryKey: [...] })` внутри самого хука
  мутации, а не в компоненте.

## Слой стейта: Zustand

- Zustand — **только** для глобального UI-стейта: фильтры, открытые
  модалки, выбранный маркетплейс, сортировка и т.п.
- Zustand **не** хранит серверные данные (карточки, цены) — это зона
  TanStack Query. Не дублировать один и тот же стейт в обоих местах.
- Все сторы живут в одном файле `lib/store.ts` — несколько независимых
  `create(...)`-хуков в одном месте, без распыления по папкам. Если файл
  станет слишком большим (4+ независимых стора), тогда переходить на
  `lib/store/<domain>.ts`, но не раньше.

```ts
// lib/store.ts
interface FilterState {
  marketplace: Marketplace | "all";
  setMarketplace: (m: Marketplace | "all") => void;
}

export const useFilterStore = create<FilterState>((set) => ({
  marketplace: "all",
  setMarketplace: (marketplace) => set({ marketplace }),
}));

interface ModalState {
  activeModal: "card-detail" | "filters" | null;
  openModal: (modal: ModalState["activeModal"]) => void;
  closeModal: () => void;
}

export const useModalStore = create<ModalState>((set) => ({
  activeModal: null,
  openModal: (activeModal) => set({ activeModal }),
  closeModal: () => set({ activeModal: null }),
}));
```

- Локальный стейт одного компонента (открыт/закрыт аккордеон и т.д.) — обычный
  `useState`, в Zustand не выносить.

## Стиль кода и нейминг

- **Файлы-компоненты** — `PascalCase.tsx` (`CardItem.tsx`, `PriceChart.tsx`).
- **Не-компонентные файлы** (хуки, утилиты, api) — `camelCase.ts`
  (`useCards.ts`, `formatPrice.ts`).
- **shadcn-компоненты** в `components/ui/*` остаются как сгенерированы
  (kebab-case файл, PascalCase экспорт) — это исключение, их не переименовываем.
- Имя компонента = имя файла без расширения.
- Каждый компонент — `export function ComponentName(...)`, без
  `export default` (кроме `page.tsx`/`layout.tsx`, где default обязателен
  по конвенции Next.js).
- Пропсы компонента — `interface ComponentNameProps`, объявлен прямо над
  компонентом в том же файле.
- Хуки всегда начинаются с `use` и возвращают объект (не массив), если
  возвращается больше одного значения.

## TypeScript

- Без `any`. Если тип неизвестен — `unknown` + явная проверка/сужение.
- Все типы домена (`Card`, `PriceEntry`, `Marketplace`) — в `lib/types.ts`,
  не дублировать инлайн в компонентах. Если файл сильно разрастётся —
  можно разбить на `lib/types/<domain>.ts`, но не раньше необходимости.
- Пропсы и возвраты функций типизируются явно, не полагаемся только на
  инференс в публичных API хуков/функций.

## Tailwind / shadcn

- Кастомные стили — только через Tailwind-классы, инлайн `style={{}}`
  не используется без крайней необходимости (динамические значения типа
  цвета по data).
- Если один и тот же набор классов повторяется 3+ раза — выносить в
  компонент, а не в `cn()`-хелпер на каждый случай.
- Новые элементы UI — сначала смотреть, нет ли готового компонента в
  shadcn, и только потом писать кастомный.
- Условные классы — через `cn()` из `lib/utils.ts`, не строковая
  конкатенация.

## Мультиязычность (KO / EN)

Сайт работает на двух языках: **корейский (основной рынок) и английский**.
Используем **next-intl** — стандартное решение для App Router.

- Все роуты живут под `app/[locale]/...`, локаль — обязательный динамический
  сегмент. Без локали в пути роутов не бывает.
- Конфигурация:
  - `i18n/routing.ts` — список локалей (`ko`, `en`) и `defaultLocale: "ko"`.
  - `i18n/request.ts` — подгрузка словаря для текущей локали.
  - `proxy.ts` (next-intl middleware) — определение/редирект локали по
    `Accept-Language` и cookie.
- Тексты — **никогда** хардкодом в JSX. Любая видимая пользователю строка
  идёт через словарь:

```ts
// messages/ko.json
{ "cards": { "title": "포켓몬 카드", "empty": "카드를 찾을 수 없습니다" } }

// messages/en.json
{ "cards": { "title": "Pokémon Cards", "empty": "No cards found" } }
```

```tsx
// в Server Component
import { getTranslations } from "next-intl/server";
const t = await getTranslations("cards");
<h1>{t("title")}</h1>

// в Client Component
import { useTranslations } from "next-intl";
const t = useTranslations("cards");
<h1>{t("title")}</h1>
```

- Namespace в словаре = домен/страница (`cards`, `marketplace`, `landing`),
  без одного гигантского плоского файла на все строки.
- Числа/цены и даты форматируются через `useFormatter` (`next-intl`), а не
  вручную через `toLocaleString` — так гарантируется консистентность между
  `ko` и `en` (разделители тысяч, форматы дат и т.п.).
- Корейская вона (₩) и числовые форматы — на стороне backend ответ должен
  приходить как число, форматирование под локаль — задача фронтенда.

## Чего не делать

- Не вызывать `fetch` напрямую в компонентах — только через
  `lib/api` → `lib/queries`.
- Не хранить серверные данные в Zustand.
- Не класть бизнес-логику в `components/ui/*`.
- Не использовать `default export` для обычных компонентов.
- Не вводить новый паттерн стейт-менеджмента (Redux, Context для
  серверных данных и т.п.) без явного запроса.
- Не хардкодить видимые пользователю строки в JSX — только через словарь
  `next-intl` (`messages/ko.json` / `en.json`).
- Не добавлять роуты вне `app/[locale]/...` — без locale-сегмента
  страница не будет правильно локализована.
- Авторизация/роли пока не реализованы — не добавлять заглушки auth-логики,
  если явно не попросили (фича запланирована на будущее).

## Когда я прошу новую фичу

При добавлении новой страницы/фичи — придерживаться этой схемы:

1. Типы в `lib/types.ts`.
2. API-функция в `lib/api/<domain>.ts`.
3. Query/Mutation-хук в `lib/queries/<domain>.ts`.
4. Стор в `lib/store.ts` — только если нужен глобальный UI-стейт.
5. Компоненты в `components/<domain>/` (переиспользуемые) или
   `components/features/<feature-name>/` (одноразовая секция страницы).
6. Тексты — в `messages/ko.json` и `messages/en.json`, не в JSX.
7. Страница в `app/[locale]/<route>/page.tsx`, минимальная — собирает
   компоненты, без логики данных внутри.