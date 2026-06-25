#!/bin/bash
# Lưu lại thay đổi
git add .
git commit -m "Cập nhật dữ liệu/code: $(date +'%d/%m/%Y %H:%M:%S')"

# Đẩy lên GitHub
git push origin main

echo "✅ Đã đẩy code và dữ liệu lên GitHub thành công!"
