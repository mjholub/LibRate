// Package meta Code generated by swaggo/swag. DO NOT EDIT
package meta

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "MJH",
            "email": "TODO@flagship.instance"
        },
        "license": {
            "name": "GNU Affero General Public License v3",
            "url": "https://www.gnu.org/licenses/agpl-3.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/artists/by-name": {
            "post": {
                "description": "Retrieve the artists with the given names",
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "media",
                    "artists",
                    "bulk operations"
                ],
                "summary": "Retrieve artists",
                "parameters": [
                    {
                        "type": "array",
                        "items": {
                            "type": "string"
                        },
                        "collectionFormat": "csv",
                        "description": "Artist names",
                        "name": "names",
                        "in": "formData",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/handlers.ResponseHTTP"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "$ref": "#/definitions/models.GroupedArtists"
                                        }
                                    }
                                }
                            ]
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/handlers.ResponseHTTP"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/handlers.ResponseHTTP"
                        }
                    }
                }
            }
        },
        "/authenticate/login": {
            "post": {
                "description": "Create a session for the user",
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth",
                    "accounts"
                ],
                "summary": "Login to the application",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Member name. Request must include either membername or email",
                        "name": "membername",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Email address",
                        "name": "email",
                        "in": "query"
                    },
                    {
                        "maximum": 2147483647,
                        "minimum": 1,
                        "type": "integer",
                        "default": 30,
                        "description": "Session time in minutes",
                        "name": "session_time",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Password",
                        "name": "password",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Referrer-Policy header",
                        "name": "Referrer-Policy",
                        "in": "header"
                    },
                    {
                        "type": "string",
                        "description": "X-CSRF-Token header",
                        "name": "X-CSRF-Token",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/handlers.ResponseHTTP"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "$ref": "#/definitions/auth.SessionResponse"
                                        }
                                    }
                                }
                            ]
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/handlers.ResponseHTTP"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/handlers.ResponseHTTP"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/handlers.ResponseHTTP"
                        }
                    }
                }
            }
        },
        "/genre/{kind}/{genre}": {
            "get": {
                "description": "Retrieve the genre with the given name and type",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "media",
                    "genres"
                ],
                "summary": "Retrieve genre",
                "parameters": [
                    {
                        "enum": [
                            "film",
                            "tv",
                            "music",
                            "book",
                            "game"
                        ],
                        "type": "string",
                        "description": "Genre kind",
                        "name": "kind",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Genre name (snake_lowercase)",
                        "name": "genre",
                        "in": "path",
                        "required": true
                    },
                    {
                        "enum": [
                            "en",
                            "de"
                        ],
                        "type": "string",
                        "description": "ISO-639-1 language code",
                        "name": "lang",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/handlers.ResponseHTTP"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "$ref": "#/definitions/models.Genre"
                                        }
                                    }
                                }
                            ]
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/handlers.ResponseHTTP"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/handlers.ResponseHTTP"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/handlers.ResponseHTTP"
                        }
                    }
                }
            }
        },
        "/genres/{kind}": {
            "get": {
                "description": "Retrieve the list of genres of the specified type",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "media",
                    "genres",
                    "bulk operations"
                ],
                "summary": "Retrieve genres",
                "parameters": [
                    {
                        "enum": [
                            "film",
                            "tv",
                            "music",
                            "book",
                            "game"
                        ],
                        "type": "string",
                        "description": "Genre kind",
                        "name": "kind",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "boolean",
                        "description": "Return only genre names. Usually used for populating dropdowns",
                        "name": "names_only",
                        "in": "query"
                    },
                    {
                        "type": "boolean",
                        "description": "Return the genre names as links",
                        "name": "as_links",
                        "in": "query"
                    },
                    {
                        "type": "boolean",
                        "description": "Return all genres, not only the ones without a parent genre (e.g. Twee Pop and Jangle Pop instead of just Pop)",
                        "name": "all",
                        "in": "query"
                    },
                    {
                        "type": "array",
                        "items": {
                            "enum": [
                                "name",
                                "id",
                                "kinds",
                                "parent",
                                "children"
                            ],
                            "type": "string"
                        },
                        "collectionFormat": "csv",
                        "description": "Return only the specified columns",
                        "name": "columns",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "If names_only=false and as_links=false",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/handlers.ResponseHTTP"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "type": "array",
                                            "items": {
                                                "$ref": "#/definitions/models.Genre"
                                            }
                                        }
                                    }
                                }
                            ]
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/handlers.ResponseHTTP"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/handlers.ResponseHTTP"
                        }
                    }
                }
            }
        },
        "/media/{id}": {
            "get": {
                "description": "Retrieve complete media information for the given media ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "media",
                    "metadata"
                ],
                "summary": "Retrieve media information",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Media UUID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/handlers.ResponseHTTP"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "type": "object"
                                        }
                                    }
                                }
                            ]
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/handlers.ResponseHTTP"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/handlers.ResponseHTTP"
                        }
                    }
                }
            }
        },
        "/media/{media_id}/cast": {
            "get": {
                "description": "Get the full cast and crew involved with the creation of the media with given ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "media",
                    "artists",
                    "bulk operations",
                    "films",
                    "television",
                    "anime"
                ],
                "summary": "Get the cast of the media with given ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The UUID of the media to get the cast of",
                        "name": "media_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/handlers.ResponseHTTP"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "$ref": "#/definitions/models.Cast"
                                        }
                                    }
                                }
                            ]
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/handlers.ResponseHTTP"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/handlers.ResponseHTTP"
                        }
                    }
                }
            }
        },
        "/media/{media_id}/images": {
            "get": {
                "description": "Retrieve the image paths for the media with the given ID",
                "consumes": [
                    "json text/plain"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "media",
                    "metadata",
                    "images"
                ],
                "summary": "Retrieve image paths",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Media UUID",
                        "name": "media_id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Media UUID",
                        "name": "media_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Image path",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/handlers.ResponseHTTP"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/handlers.ResponseHTTP"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/handlers.ResponseHTTP"
                        }
                    }
                }
            }
        },
        "/members/{email_or_username}/info": {
            "get": {
                "description": "Retrieve the information the requester is allowed to see about a member",
                "consumes": [
                    "json application/activity+json"
                ],
                "tags": [
                    "accounts",
                    "interactions",
                    "metadata"
                ],
                "summary": "Get a member (user) by nickname or email",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The nickname or email of the member to get",
                        "name": "email_or_username",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/handlers.ResponseHTTP"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "$ref": "#/definitions/member.Member"
                                        }
                                    }
                                }
                            ]
                        }
                    },
                    "401": {
                        "description": "When certain access prerequisites are not met, e.g. a follower's only-visible metadata is requested",
                        "schema": {
                            "$ref": "#/definitions/handlers.ResponseHTTP"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/handlers.ResponseHTTP"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/handlers.ResponseHTTP"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "auth.SessionResponse": {
            "type": "object",
            "properties": {
                "memberName": {
                    "type": "string",
                    "example": "lain"
                },
                "token": {
                    "type": "string",
                    "example": "[A-Za-z0-9]{37}.[A-Za-z0-9]{147}.L-[A-Za-z0-9]{24}_[A-Za-z0-9]{25}-zNjCwGMr-[A-Za-z0-9]{27}"
                }
            }
        },
        "handlers.ResponseHTTP": {
            "type": "object",
            "properties": {
                "data": {},
                "message": {
                    "type": "string"
                },
                "status": {
                    "type": "integer"
                }
            }
        },
        "member.Member": {
            "type": "object",
            "required": [
                "email",
                "memberName"
            ],
            "properties": {
                "active": {
                    "type": "boolean",
                    "example": true
                },
                "bio": {
                    "type": "string",
                    "example": "Wherever you go, everyone is connected."
                },
                "displayName": {
                    "type": "string",
                    "example": "Lain Iwakura"
                },
                "email": {
                    "type": "string",
                    "example": "lain@wired.jp"
                },
                "followers_uri": {
                    "description": "URI for getting the followers list of this account",
                    "type": "string"
                },
                "following_uri": {
                    "description": "URI for getting the following list of this account",
                    "type": "string"
                },
                "homepage": {
                    "type": "string",
                    "example": "https://webnavi.neocities.org/"
                },
                "irc": {
                    "description": "doomed fields, will be removed by arbitrary user-defined fields",
                    "type": "string"
                },
                "matrix": {
                    "type": "string"
                },
                "memberName": {
                    "description": "MemberName != webfinger",
                    "type": "string",
                    "maxLength": 30,
                    "minLength": 3,
                    "example": "lain"
                },
                "profile_pic": {
                    "type": "string",
                    "example": "/static/img/profile/lain.jpg"
                },
                "regdate": {
                    "type": "string",
                    "example": "2020-01-01T00:00:00Z"
                },
                "roles": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "[\"admin\"",
                        " \"moderator\"]"
                    ]
                },
                "uuid": {
                    "type": "string"
                },
                "visibility": {
                    "type": "string",
                    "example": "followers_only"
                },
                "xmpp": {
                    "type": "string"
                }
            }
        },
        "models.Cast": {
            "type": "object",
            "properties": {
                "ID": {
                    "type": "integer"
                },
                "actors": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.Person"
                    }
                },
                "directors": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.Person"
                    }
                }
            }
        },
        "models.Country": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "name": {
                    "type": "string"
                }
            }
        },
        "models.Genre": {
            "type": "object",
            "properties": {
                "children": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                },
                "description": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.GenreDescription"
                    }
                },
                "id": {
                    "type": "integer"
                },
                "keywords": {
                    "description": "DescLong    string   ` + "`" + `json:\"desc_long\" db:\"desc_long\"` + "`" + `",
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "['dark'",
                        " 'gloomy'",
                        " 'atmospheric'",
                        " 'raw'",
                        " 'underproduced']"
                    ]
                },
                "kind": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "music"
                    ]
                },
                "name": {
                    "type": "string",
                    "example": "Black Metal"
                },
                "parent_genre": {
                    "type": "integer"
                }
            }
        },
        "models.GenreDescription": {
            "type": "object",
            "properties": {
                "description": {
                    "type": "string",
                    "example": "Typified by highly distorted, trebly, tremolo-picked guitars, blast beats, double kick drumming, shrieked vocals, and raw, underproduced sound that often favors atmosphere over technical skills and melody."
                },
                "genre_id": {
                    "type": "integer",
                    "example": 2958
                },
                "language": {
                    "type": "string",
                    "example": "en"
                }
            }
        },
        "models.Group": {
            "type": "object",
            "properties": {
                "active": {
                    "type": "boolean"
                },
                "added": {
                    "type": "string"
                },
                "bandcamp": {
                    "type": "string"
                },
                "bio": {
                    "type": "string"
                },
                "disbanded": {
                    "type": "string"
                },
                "formed": {
                    "type": "string"
                },
                "genres": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.Genre"
                    }
                },
                "id": {
                    "type": "string"
                },
                "kind": {
                    "description": "Orchestra, Choir, Ensemble, Collective, etc.",
                    "type": "string"
                },
                "locations": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.Place"
                    }
                },
                "members": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.Person"
                    }
                },
                "modified": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "photos": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "primary_genre": {
                    "$ref": "#/definitions/models.Genre"
                },
                "soundcloud": {
                    "type": "string"
                },
                "website": {
                    "type": "string"
                },
                "wikipedia": {
                    "type": "string"
                },
                "works": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "models.GroupedArtists": {
            "type": "object",
            "properties": {
                "group": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.Group"
                    }
                },
                "individual": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.Person"
                    }
                }
            }
        },
        "models.Person": {
            "type": "object",
            "properties": {
                "added": {
                    "type": "string"
                },
                "bio": {
                    "type": "string",
                    "example": "wojtyła disco dance"
                },
                "birth": {
                    "description": "DOB can also be unknown",
                    "type": "string"
                },
                "death": {
                    "type": "string",
                    "example": "2005-04-02T21:37:00Z"
                },
                "first_name": {
                    "type": "string",
                    "example": "Karol"
                },
                "hometown": {
                    "$ref": "#/definitions/models.Place"
                },
                "id": {
                    "type": "string",
                    "example": "12345678-90ab-cdef-9876-543210fedcba"
                },
                "last_name": {
                    "type": "string",
                    "example": "Wojtyła"
                },
                "modified": {
                    "type": "string"
                },
                "name": {
                    "description": "helper field for complete name",
                    "type": "string"
                },
                "nick_names": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "['pawlacz'",
                        " 'jan pawulon']"
                    ]
                },
                "other_names": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "['Jan Paweł II']"
                    ]
                },
                "photos": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "residence": {
                    "$ref": "#/definitions/models.Place"
                },
                "roles": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "website": {
                    "type": "string",
                    "example": "https://www.vatican.va/content/john-paul-ii/en.html"
                },
                "works": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "models.Place": {
            "type": "object",
            "properties": {
                "country": {
                    "$ref": "#/definitions/models.Country"
                },
                "id": {
                    "type": "integer"
                },
                "kind": {
                    "type": "string"
                },
                "lat": {
                    "type": "number"
                },
                "lng": {
                    "type": "number"
                },
                "name": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "dev",
	Host:             "",
	BasePath:         "/api",
	Schemes:          []string{},
	Title:            "LibRate",
	Description:      "API for LibRate, a social media cataloguing and reviewing service",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}