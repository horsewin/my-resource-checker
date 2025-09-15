# AWS Resource Validator for SBCNTR Hands-on

AWS Resource Validatorは、SBCNTRハンズオンで作成されるAWSリソースを検証するGolang製のCLIツールです。AWS Cloud Control APIを使用して、各ステップでリソースが正しく作成・設定されているかを自動的にチェックします。

## インストール

### 前提条件

- Go 1.23以上
- AWS CLIが設定済み（認証情報）
- 適切なIAMポリシー（読み取り権限のみを与えること）

### ビルド方法

```bash
# リポジトリのクローン
git clone https://github.com/your-org/sbcntr2-test-tool.git
cd sbcntr2-test-tool

# 依存関係のインストール
go mod tidy

# ビルド
go build -o sbcntr-validator

# 実行権限の付与
chmod +x sbcntr-validator
```

## 使い方

### 基本的な使用方法

```bash
# 特定のステップを検証
./sbcntr-validator validate --step 1

# 全ステップを検証
./sbcntr-validator validate --all

# 詳細出力モード
./sbcntr-validator validate --step 1 --verbose

# JSON形式で出力
./sbcntr-validator validate --all --output json

# 特定のAWSプロファイルを使用
./sbcntr-validator validate --step 1 --profile myprofile

# 特定のリージョンを指定
./sbcntr-validator validate --step 1 --region ap-northeast-1
```

### コマンドオプション

| オプション | 短縮形 | 説明 | デフォルト |
|---------|--------|------|----------|
| `--step` | `-s` | 検証するステップ番号（1-6） | - |
| `--all` | `-a` | 全ステップを検証 | false |
| `--verbose` | `-v` | 詳細情報を表示 | false |
| `--output` | `-o` | 出力形式（console/json） | console |
| `--profile` | `-p` | AWS プロファイル名 | default |
| `--region` | `-r` | AWS リージョン | ap-northeast-1 |
| `--config` | | 設定ファイルのパス | ~/.sbcntr-validator.yaml |

## ステップ概要

### Step 1: ネットワーク構築
- VPC、サブネット、セキュリティグループ、インターネットゲートウェイの検証
- 書籍における【XXX節：コンテナレジストリの構築】の前までの状態を検証

### Step 2: ECRリポジトリセットアップ
- フロントエンド/バックエンド用のECRリポジトリの検証
- 書籍における【XXX節：コンテナレジストリの作成】後の状態を検証

### Step 3: VPCエンドポイント設定
- ECR API、ECR DKR、S3、CloudWatch Logs用のVPCエンドポイントの検証
- 書籍における【XXX節：Blue/Green デプロイメント用の ALB 追加】の前までの状態を検証

### Step 4: ECSクラスターとロードバランサー
- ECSクラスター、ALB、ターゲットグループの検証
- EcsInfrastructureRoleForLoadBalancers ロールの存在確認
- AmazonECSInfrastructureRolePolicyForLoadBalancersポリシーのアタッチメント確認
- 書籍における【XXX節：ECSの構築】までの状態を検証

### Step 5: ECSサービスデプロイ
- タスク定義とECSサービスの検証

### Step 6: データベース構成
- Aurora DBクラスターとインスタンスの検証
- DBサブネットグループの設定確認
- DB用セキュリティグループの検証（バックエンドからのアクセス許可）
- マスターユーザー名とエンジンタイプの確認

## 必要なIAMポリシー

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "cloudcontrol:GetResource",
        "cloudcontrol:ListResources",
        "cloudformation:DescribeStacks",
        "cloudformation:ListStackResources",
        "ec2:DescribeVpcs",
        "ec2:DescribeSubnets",
        "ec2:DescribeSecurityGroups",
        "ec2:DescribeInternetGateways",
        "ec2:DescribeVpcEndpoints",
        "ecs:DescribeClusters",
        "ecs:DescribeServices",
        "ecs:DescribeTaskDefinition",
        "ecs:ListClusters",
        "ecr:DescribeRepositories",
        "elasticloadbalancing:DescribeLoadBalancers",
        "elasticloadbalancing:DescribeTargetGroups",
        "iam:GetRole",
        "iam:ListAttachedRolePolicies"
      ],
      "Resource": "*"
    }
  ]
}
```

## 出力例

### コンソール出力

```
============================================================
STEP 1: Network Construction
============================================================
Status: ✅ PASSED
Duration: 2.34s

Resources Checked:
----------------------------------------
✅ sbcntr (AWS::EC2::VPC)
✅ sbcntr-subnet-public-ingress-1a (AWS::EC2::Subnet)
✅ sbcntr-subnet-public-ingress-1c (AWS::EC2::Subnet)
✅ sbcntr-sg-ingress (AWS::EC2::SecurityGroup)

============================================================
✅ All checks passed! You can proceed to the next step.
```

### JSON出力

```json
{
  "stepNumber": 1,
  "stepName": "Network Construction",
  "status": "PASSED",
  "duration": "2.34s",
  "resources": [
    {
      "type": "AWS::EC2::VPC",
      "id": "vpc-12345",
      "name": "sbcntr",
      "status": "EXISTS",
      "expected": {
        "CidrBlock": "10.0.0.0/16"
      },
      "actual": {
        "CidrBlock": "10.0.0.0/16"
      }
    }
  ]
}
```

## トラブルシューティング

### よくあるエラーと解決方法

1. **認証エラー**
   ```
   Error: failed to initialize AWS client
   ```
   解決方法: AWS CLIの認証設定を確認してください。

2. **リソースが見つからない**
   ```
   ❌ Required resource 'sbcntr' not found
   ```
   解決方法: ハンズオンの該当ステップを実行済みか確認してください。

3. **設定値の不一致**
   ```
   ⚠️ VPC CIDR block should be 10.0.0.0/16: expected 10.0.0.0/16, got 172.16.0.0/16
   ```
   解決方法: リソースの設定値を修正してください。

## 開発

### テストの実行

```bash
# 単体テスト
go test ./...

# カバレッジレポート
go test -cover ./...

# 統合テスト（AWS環境必要）
go test -tags=integration ./...
```

### ビルド（クロスコンパイル）

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o sbcntr-validator-linux

# macOS
GOOS=darwin GOARCH=amd64 go build -o sbcntr-validator-darwin

# Windows
GOOS=windows GOARCH=amd64 go build -o sbcntr-validator.exe
```

## ライセンス

MIT License

## コントリビューション

プルリクエストを歓迎します。大きな変更の場合は、まずissueを開いて変更内容を議論してください。

## サポート

問題が発生した場合は、[GitHubのIssue](https://github.com/your-org/sbcntr2-test-tool/issues)で報告してください。