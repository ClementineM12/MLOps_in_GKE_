import os 
from glob import glob 
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
from sklearn.utils import resample
import cv2

import PIL.Image as Image
import numpy as np


# Creating dictionary for displaying more human-friendly labels
lesion_type_dict = {
    'nv': 'Melanocytic nevi',
    'mel': 'Melanoma',
    'bkl': 'Benign keratosis-like lesions',
    'bcc': 'Basal cell carcinoma',
    'akiec': 'Actinic keratoses',
    'vasc': 'Vascular lesions',
    'df': 'Dermatofibroma'
}

def process(skin_meta):

    # Merge images from both folders into one dictionary
    imageid_path_dict = {os.path.splitext(os.path.basename(x))[0]: x for x in glob(os.path.join('*', '*.jpg'))}

    # Creating new columns for better understanding of features
    skin_meta['path'] = skin_meta['image_id'].map(imageid_path_dict.get)
    skin_meta['Cell type'] = skin_meta['dx'].map(lesion_type_dict.get)
    skin_meta['target'] = pd.Categorical(skin_meta['Cell type']).codes

    skin_meta.head()

    # Look for missing data
    skin_meta.isna().sum()
    # Replace with the mean 
    skin_meta['age'].fillna((skin_meta['age'].mean()), inplace=True)

    plot_1(skin_meta)
    plot_2(skin_meta)

    skin_meta = skin_meta.assign(label = skin_meta.apply(lesion_type, axis=1))

    skin_meta['image'] = skin_meta['path'].map(lambda x: np.asarray(Image.open(x)))
    balanced_data = create_balanced_dataset(skin_meta)

    create_folder_segmented()
    create_segmented_pictures(balanced_data['image'], balanced_data, "Segmented_images")

def plot_1(skin_meta):
    # Barplot with Counts per cell Types of Skin lesions
    sns.set_style('darkgrid', {"grid.color": ".8", "grid.linestyle": ":"})
    fig,axes = plt.subplots(figsize=(12,8))

    ax = sns.countplot(x = 'Cell type', data = skin_meta, palette = 'coolwarm')
    for container in ax.containers:
        ax.bar_label(container)
    plt.title('Counts per cell Types of Skin lesions')
    plt.ylabel('Counts')
    plt.xlabel('Skin lesion type')
    plt.xticks(rotation=35)
    plt.show()


def plot_2(skin_meta):
    # Skin lesion type occurence by Age
    sns.set_style('darkgrid', {"grid.color": ".8", "grid.linestyle": ":"})
    fig,axes = plt.subplots(figsize=(12,8))
    ax = sns.histplot(data = skin_meta, x = 'age', hue = 'Cell type', multiple='stack', palette='coolwarm')
    plt.title('Skin lesion type occurrence by Age')
    plt.ylabel('Counts')
    plt.xlabel('Age')
    plt.show()


# Create a column 'label' with categories of malignant and benign skin lesions
def lesion_type(row):
    if row['dx'] == 'bkl' or row['dx'] == 'df' or row['dx'] == 'nv' or row['dx'] == 'vasc' :           
        return 0 # 'benign'
    elif row['dx'] == 'akiec' or row['dx'] == 'mel' or row['dx'] == 'bcc' :
        return 1 # 'malignant'


def create_balanced_dataset(skin_meta, n_samples=2000, seed=42):
    # Create a balanced dataset to use later in the models
    skin_0 = skin_meta[skin_meta['label'] == 0]
    skin_1 = skin_meta[skin_meta['label'] == 1]

    skin_0_balanced = resample(skin_0, replace=True, n_samples=n_samples, random_state=seed) 
    skin_1_balanced = resample(skin_1, replace=True, n_samples=n_samples, random_state=seed) 

    skin_200 = pd.concat([skin_0_balanced, skin_1_balanced])
    skin_200 = skin_200.reset_index()

    shape = skin_meta['image'][0].shape
    print(f"Image shape: {shape}")

    return skin_200


def create_folder_segmented():
    # Create a folder in the directory to save inside the segmented images
    folder_name = "Segmented_images"
    if not os.path.exists(folder_name):
        os.makedirs(folder_name)


def create_segmented_pictures(images_array, skin_200, folder_name):
    kernel = np.ones((8,8), np.uint8)
    skin_200['segmented_image'] = ''
    for i in range(len(images_array)):
        image_gr = cv2.cvtColor(images_array[i], cv2.COLOR_BGR2GRAY)
        image_blur = cv2.medianBlur(image_gr, 5)
        otsu_threshold, image_result = cv2.threshold(image_blur, 0, 255, cv2.THRESH_BINARY + cv2.THRESH_OTSU)
        image_cropped = image_result[60:400, 50:550]
        image_cropped = cv2.bitwise_not(image_cropped)

        # Save the image in the file "Segmented_images"
        skin_200['segmented_image'][i] = np.array(image_cropped).astype(np.uint8)
        
        image_saved = Image.fromarray(image_cropped)
        image_seg_id = skin_200['image_id'][i]
        image_saved.save(f'{folder_name}/segmented_{image_seg_id}.jpg', 'JPEG')
        