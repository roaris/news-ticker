# news-ticker

毎朝8時ごろにニュースを送信するLINE Botです

<img src="https://user-images.githubusercontent.com/61813626/157556323-fdac5afb-a7d2-4f44-bdc1-78f4eee51fe7.png" width=800>

## 友だち追加

下の QR コードから友だち追加できます

<img src="https://user-images.githubusercontent.com/61813626/157553123-7f3de2a6-3571-4762-813f-53d6cb2ffac3.png" width=250 />

## 使い方
- Bot から送信するニュースのカテゴリを選ぶことができます
- カテゴリはビジネス、エンタメ、健康、科学、スポーツ、テクノロジーから選ぶことができます
- カテゴリを複数選ぶことが可能です
- +カテゴリ名と送信すると、カテゴリの追加、-カテゴリ名と送信すると、カテゴリの削除が行われます(例. +ビジネス、-エンタメ)
- デフォルトのカテゴリはビジネスになっています
- 登録されているカテゴリがない状態にすると、ニュースの送信を停止します

## 使用技術
- Go
- SAM, Lambda, DynamoDB
- LINE Messaging API
- News API

## アーキテクチャ
<img src="https://user-images.githubusercontent.com/61813626/157560438-8bc50dce-3968-422c-a2d3-ac41910164d2.png" width=800>
