// app.js — Client-side JavaScript for Product Catalog

function showToast(message, type) {
    var toast = document.getElementById('toast');
    toast.textContent = message;
    toast.className = 'toast toast-' + type;
    void toast.offsetWidth;
    toast.classList.add('toast-visible');

    setTimeout(function() {
        toast.classList.remove('toast-visible');
        setTimeout(function() {
            toast.className = 'toast hidden';
        }, 300);
    }, 3000);
}

async function deleteProduct(id) {
    if (!confirm('Are you sure you want to delete this product?')) {
        return;
    }

    try {
        var response = await fetch('/products/' + id, {
            method: 'GET',
        });
        if (response.ok) {
            showToast('Deleted!', 'success');
            setTimeout(function() {
                window.location.href = '/';
            }, 1000);
        } else {
            showToast('Failed to delete product', 'error');
        }
    } catch (err) {
        showToast('Network error', 'error');
    }
}

async function purchaseProduct(id) {
    try {
        var response = await fetch('/products/' + id + '/purchase', {
            method: 'POST',
        });
        if (response.ok) {
            showToast('Purchased!', 'success');
            var qtyEl = document.getElementById('product-quantity');
            if (qtyEl) {
                var current = parseInt(qtyEl.textContent);
                qtyEl.textContent = Math.max(0, current - 1);
            }
        } else {
            var data = await response.json();
            showToast(data.error || 'Purchase failed', 'error');
        }
    } catch (err) {
        showToast('Network error', 'error');
    }
}

async function createProduct(event) {
    event.preventDefault();
    var form = event.target;
    var formData = new FormData(form);

    var data = {
        name: formData.get('name') || '',
        description: formData.get('description') || '',
        price: parseFloat(formData.get('price')) || 0,
        category: formData.get('category') || '',
        quantity: parseInt(formData.get('quantity')) || 0,
        in_stock: (parseInt(formData.get('quantity')) || 0) > 0
    };

    try {
        var response = await fetch('/products', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data),
        });
        if (response.ok) {
            showToast('Product created!', 'success');
            setTimeout(function() {
                window.location.href = '/';
            }, 1000);
        }
    } catch (err) {
        document.getElementById('error-msg').style.display = 'block';
    }
}
