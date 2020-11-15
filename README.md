# API Project

## Goal

Simple project API for Timed Series Event
CRUD

### DataSet Exemple

```cs
class Event
{
    private int id
    private timestamptz time;
    private int[] flags ;
    private json data;
}
```

Stored in SQL db (ideallly postgres, else mariadb)

## Specifications

### Validité des données

time ne peut pas être null. Si non fourni, insérer avec NOW()

flags ne peut pas être null, il doit contenir au moins une valeur

data ne peut pas être null. Il peut contenin un json vide.

### Structure du JSON

```json
{
    "id":    int,
    "time":  timestamptz,
    "flags": int[],
    "data":  string // json
}
```

#### Valid Json example

```json
{
    "id": 5,
    "time": "2020-11-13 20:01:49.849163+01",
    "flags":[1,12,3],
    "data":"{\"timer_niveau4\": 1254}"
}
```

### API Specs

#### Global

- si succès : 200 (OK) (+data)
- si les paramètre ne sont pas déserialisables : 422 (Unprocessable Entity)
- si aucune entrée avec cette id : 404 (Not Found)
- si erreur backend SQL : 500 (internal Server Error) + (SQL Error)

## routes

GET /api/{controller}\
Renvoie tout les events. LIMIT par défaut de 500 lignes.

GET /api/{controller}/{id}\
Renvoie l'event correspondant à id.

GET /api/{controller}/getFlag/{flag}\
Renvoie l'ensemble des events contenant flag. LIMIT par défaut de 500 lignes.

POST /api/{controller}\
Insère un ou plusieurs events dans la base. Renvoie le nombre de ligne impactées.

PATCH /api/{controller}\
Met à jour un ou plusieurs events. Renvoie le nombre de lignes impactées.

DELETE /api/{controller}/{id}\
Met tous les champs de l'event id à la valeur neutre (id = $id, time = EPOCH, flags=[-1], data="{}"). Renvoie l'id impactée.

DELETE /api/{controller}/deleteflag/{flag}\
Met tout les champs des évènements contenant flag à la valeur neutre (id = $id, time = EPOCH, flags=[-1], data="{}"). Renvoie le nombre de lignes impactées.

## Env variable to define

APP_PORT
