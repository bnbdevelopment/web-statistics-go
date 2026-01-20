# Web Statisztikák Go

Ez a projekt egy webstatisztikai alkalmazás, amely a Go nyelven íródott backendből és egy Next.js frontendből áll. Lehetővé teszi a webhely forgalmának nyomon követését, beleértve az egyedi látogatókat, a munkameneteket, a földrajzi helyzetet és egyebeket.

## Telepítés

A projektet Docker és Docker Compose segítségével lehet a legegyszerűbben futtatni.

### Előfeltételek

-   [Docker](https://docs.docker.com/get-docker/)
-   [Docker Compose](https://docs.docker.com/compose/install/)

### Telepítési lépések

1.  **Klónozza a repository-t:**

    ```bash
    git clone https://github.com/gemini-testing/statistics_go.git
    cd statistics_go
    ```

2.  **Hozzon létre egy `.env` fájlt:**

    Másolja át a `.env.example` fájlt `.env` néven, és állítsa be a környezeti változókat.

    ```bash
    cp .env.example .env
    ```

    A `.env` fájlban a következő változókat kell beállítani:

    ```bash
    # Backend
    BACKEND_PORT=3001
    GIN_MODE=release
    PREFIX=/api

    # TimescaleDB
    POSTGRES_DB=timescaledb
    POSTGRES_USER=root
    POSTGRES_PASSWORD=12345
    ```

3.  **Indítsa el az alkalmazást a Docker Compose segítségével:**

    ```bash
    docker-compose up -d --build
    ```

    Ez a parancs felépíti a backend és a frontend image-eket, és elindítja a konténereket a háttérben.

4.  **Ellenőrizze, hogy az alkalmazás fut-e:**

    -   A frontend a `http://localhost:3000` címen érhető el.
    -   A backend a `http://localhost:3001` címen érhető el.

## Hitelesítési Proxy

Ajánlott egy hitelesítési proxyt (pl. Nginx, Traefik, Caddy) elé helyezni az alkalmazásnak, különösen éles környezetben. Ez a következő előnyökkel jár:

-   **Biztonság:** A proxy képes kezelni a HTTPS-t, a hitelesítést és más biztonsági funkciókat.
-   **Terheléselosztás:** Több backend példány esetén a proxy eloszthatja a terhelést.
-   **Egységes hozzáférés:** A frontend és a backend egyetlen domainen keresztül érhető el.

## API Végpontok

Minden API végpont a `.env` fájlban definiált `PREFIX` alatt érhető el (alapértelmezetten `/api`).

### `GET /put-traffic`

Rögzít egy felhasználói látogatást.

**Query paraméterek:**

-   `sessionId` (opcionális): A felhasználó egyedi azonosítója. Ha nem adjuk meg, a rendszer generál egy újat, és visszaküldi a válaszban.
-   `page`: A meglátogatott oldal URL-címe.
-   `site`: A meglátogatott webhely domainje.

**Példa kérés:**

```
GET /api/put-traffic?sessionId=9069c164-d8f5-4734-bb8c-72d12f6e788e&page=/&site=example.com
```

**Válasz:**

A válasz a `sessionId`-t tartalmazza.

### `POST /traffic`

Visszaadja az egyedi látogatók számát a megadott időintervallumban.

**Query paraméterek:**

-   `from`: A kezdő dátum (formátum: `YYYY-MM-DD`).
-   `to`: A záró dátum (formátum: `YYYY-MM-DD`).
-   `page`: A nyomon követett webhely.

**Válasz:**

```json
{
    "traffic": 123
}
```

### `POST /sites`

Visszaadja a különböző oldalak egyedi látogatóinak számát.

**Query paraméterek:**

-   `from`: A kezdő dátum (formátum: `YYYY-MM-DD`).
-   `to`: A záró dátum (formátum: `YYYY-MM-DD`).
-   `page`: A nyomon követett webhely.

**Válasz:**

```json
[
    {
        "page": "/",
        "users": 100
    },
    {
        "page": "/about",
        "users": 50
    }
]
```

### `POST /graph`

Visszaadja a forgalmi statisztikákat a megadott időintervallumban, `intervals` számú részre bontva.

**Query paraméterek:**

-   `from`: A kezdő dátum (formátum: `YYYY-MM-DD`).
-   `to`: A záró dátum (formátum: `YYYY-MM-DD`).
-   `page`: A nyomon követett webhely.
-   `intervals`: A szakaszok száma.

**Válasz:**

```json
[
    {
        "interval": 0,
        "uniqueSessions": 10,
        "totalRequests": 25
    },
    ...
]
```

### `POST /active`

Visszaadja az aktív felhasználók számát valós időben.

**Válasz:**

```json
{
    "count": 12
}
```

### `POST /time`

Visszaadja a felhasználók által az oldalon átlagosan eltöltött időt.

**Query paraméterek:**

-   `from`: A kezdő dátum (formátum: `YYYY-MM-DD`).
-   `to`: A záró dátum (formátum: `YYYY-MM-DD`).
-   `page`: A nyomon követett webhely.

**Válasz:**

```json
{
    "count": 120 // másodpercben
}
```

### `POST /get-sites`

Visszaadja az összes nyomon követett webhely listáját.

**Válasz:**

```json
{
    "sites": [
        "example.com",
        "another-site.com"
    ]
}
```

### `POST /get-locations`

Visszaadja a felhasználók földrajzi helyzetét.

**Query paraméterek:**

-   `from`: A kezdő dátum (formátum: `YYYY-MM-DD`).
-   `to`: A záró dátum (formátum: `YYYY-MM-DD`).
-   `page`: A nyomon követett webhely.

**Válasz:**

```json
{
    "locations": [
        {
            "latitude": 47.4979,
            "longitude": 19.0402,
            "count": 5
        },
        ...
    ]
}
```

### `GET /health`

Egészség-ellenőrző végpont.

**Válasz:**

```json
{
    "status": "healthy"
}
```
