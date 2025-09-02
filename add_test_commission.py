#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
测试佣金记录添加脚本 - Python版本
用于向MongoDB数据库添加测试佣金数据
"""

import os
import sys
from datetime import datetime, timedelta, timezone
from pymongo import MongoClient
from bson import ObjectId
from typing import List, Dict, Any
import logging

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)


class MongoDBConnection:
    """MongoDB连接管理类"""
    
    def __init__(self):
        self.client = None
        self.db = None
        
    def connect(self):
        """连接到MongoDB数据库"""
        try:
            # 从环境变量获取MongoDB URL，默认使用生产环境的认证URL
            mongodb_url = os.getenv('MONGODB_URL', 'mongodb://miniprogram_db:Chenqichen666@113.45.220.0/miniprogram_db?authSource=miniprogram_db')
            
            # 创建MongoDB客户端
            self.client = MongoClient(mongodb_url, serverSelectionTimeoutMS=10000)
            
            # 选择数据库（与Go版本保持一致）
            self.db = self.client['miniprogram_db']
            
            # 测试连接
            self.client.admin.command('ping')
            logger.info(f"成功连接到MongoDB, 地址为: {mongodb_url}")
            
        except Exception as e:
            logger.error(f"MongoDB连接失败: {e}")
            raise
            
    def get_collection(self, collection_name: str):
        """获取指定名称的集合"""
        if self.db is None:
            raise Exception("数据库未初始化")
        return self.db[collection_name]
        
    def close(self):
        """关闭数据库连接"""
        if self.client:
            self.client.close()
            logger.info("MongoDB连接已关闭")


class CommissionTestData:
    """佣金测试数据生成器"""
    
    def __init__(self, user_openid: str):
        self.user_openid = user_openid
        
    def generate_test_commissions(self) -> List[Dict[str, Any]]:
        """生成测试佣金记录"""
        now = datetime.now(timezone.utc).replace(tzinfo=None)
        
        test_commissions = [
            {
                "_id": ObjectId(),
                "commission_id": "test_commission_1",
                "user_openid": self.user_openid,
                "amount": 1.00,  # 1元佣金
                "date": now - timedelta(days=5),  # 5天前
                "status": "paid",  # 已支付状态
                "type": "agent",  # 代理佣金
                "description": "测试佣金记录1 - 代理销售佣金",
                "order_id": "test_order_001",
                "created_at": now - timedelta(days=5),
                "updated_at": now - timedelta(days=5)
            },
            {
                "_id": ObjectId(),
                "commission_id": "test_commission_2",
                "user_openid": self.user_openid,
                "amount": 0.50,  # 0.5元佣金
                "date": now - timedelta(days=3),  # 3天前
                "status": "paid",  # 已支付状态
                "type": "referral",  # 推荐佣金
                "description": "测试佣金记录2 - 推荐佣金",
                "order_id": "test_order_002",
                "referred_user_openid": "test_referred_user",
                "referred_user_name": "测试被推荐用户",
                "created_at": now - timedelta(days=3),
                "updated_at": now - timedelta(days=3)
            },
            {
                "_id": ObjectId(),
                "commission_id": "test_commission_3",
                "user_openid": self.user_openid,
                "amount": 0.25,  # 0.25元佣金
                "date": now - timedelta(days=1),  # 1天前
                "status": "paid",  # 已支付状态
                "type": "agent",  # 代理佣金
                "description": "测试佣金记录3 - 代理销售佣金",
                "order_id": "test_order_003",
                "created_at": now - timedelta(days=1),
                "updated_at": now - timedelta(days=1)
            }
        ]
        
        return test_commissions


def main():
    """主函数"""
    # 测试用户的OpenID（从调试日志中获取）
    test_user_openid = "oI0kY7byREWCaGvrN7hgWypKK-CM"
    
    # 创建数据库连接
    db_conn = MongoDBConnection()
    
    try:
        # 连接数据库
        db_conn.connect()
        
        # 获取佣金集合
        commissions_collection = db_conn.get_collection("commissions")
        
        # 生成测试数据
        test_data_generator = CommissionTestData(test_user_openid)
        test_commissions = test_data_generator.generate_test_commissions()
        
        # 插入测试数据
        logger.info("开始插入测试佣金记录...")
        
        for commission in test_commissions:
            try:
                result = commissions_collection.insert_one(commission)
                logger.info(
                    f"成功添加测试佣金记录: ID={commission['commission_id']}, "
                    f"金额={commission['amount']:.2f}元, 状态={commission['status']}"
                )
            except Exception as e:
                logger.error(f"插入佣金记录失败: {e}")
        
        # 计算总佣金
        total_amount = sum(c['amount'] for c in test_commissions)
        
        # 输出结果
        print("\n=== 测试数据添加完成 ===")
        print(f"用户OpenID: {test_user_openid}")
        print(f"添加的佣金记录数量: {len(test_commissions)}")
        print(f"总佣金金额: {total_amount:.2f}元")
        print(f"现在用户应该可以提取最多 {total_amount:.2f}元")
        print("\n请重新测试提取0.01元的API请求")
        
    except Exception as e:
        logger.error(f"脚本执行失败: {e}")
        sys.exit(1)
        
    finally:
        # 关闭数据库连接
        db_conn.close()


if __name__ == "__main__":
    main()